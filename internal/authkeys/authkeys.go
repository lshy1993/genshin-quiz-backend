// Package authkeys 管理 MoeLink 统一登录使用的 ES256 签名密钥。
//
// 只有 auth-service 持有私钥并签发 token；其他服务（quiz、node-api、gateway 等）
// 只需要公钥，通过 auth-service 暴露的 JWKS 地址获取并校验，不再共享任何签名密钥。
package authkeys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Issuer = "moelink-auth"

	jwksMinRetryGap  = 10 * time.Second
	jwksFetchTimeout = 5 * time.Second
)

var (
	ErrNoKeys       = errors.New("authkeys: no verification keys available")
	ErrInvalidToken = errors.New("authkeys: invalid token")
	ErrNotSigner    = errors.New("authkeys: this process cannot sign tokens")
)

// AccessTokenTTL 是 access token 的有效期，可用 ACCESS_TOKEN_TTL（如 15m、24h）覆盖。
func AccessTokenTTL() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_TTL")); err == nil && d > 0 {
		return d
	}
	return 15 * time.Minute
}

// Claims 是校验通过后的 token 内容。
type Claims struct {
	UserID    int64
	Email     string
	ExpiresAt time.Time
	ID        string
	// ClientID 非空表示这是发给第三方（站外）应用的受限 token，不能用于 MoeLink 自家的接口。
	ClientID string
}

type keySet struct {
	mu      sync.RWMutex
	keys    map[string]*ecdsa.PublicKey
	lastTry time.Time
}

func (s *keySet) get(kid string) (*ecdsa.PublicKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key, ok := s.keys[kid]
	return key, ok
}

func (s *keySet) replace(keys map[string]*ecdsa.PublicKey) {
	s.mu.Lock()
	s.keys = keys
	s.mu.Unlock()
}

type state struct {
	set      keySet
	signer   *ecdsa.PrivateKey
	kid      string
	jwksURL  string
	jwksJSON []byte
	client   *http.Client
}

var (
	mu      sync.RWMutex
	current *state
)

func get() *state {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

func keyID(pub *ecdsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(der)
	return base64.RawURLEncoding.EncodeToString(sum[:8]), nil
}

func padTo32(b []byte) []byte {
	if len(b) >= 32 {
		return b
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

func jwkFor(pub *ecdsa.PublicKey, kid string) map[string]string {
	return map[string]string{
		"kty": "EC",
		"crv": "P-256",
		"alg": "ES256",
		"use": "sig",
		"kid": kid,
		"x":   base64.RawURLEncoding.EncodeToString(padTo32(pub.X.Bytes())),
		"y":   base64.RawURLEncoding.EncodeToString(padTo32(pub.Y.Bytes())),
	}
}

func loadOrCreatePrivateKey(path string) (*ecdsa.PrivateKey, bool, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		block, _ := pem.Decode(data)
		if block == nil {
			return nil, false, fmt.Errorf("authkeys: %s is not valid PEM", path)
		}
		key, parseErr := x509.ParseECPrivateKey(block.Bytes)
		if parseErr != nil {
			return nil, false, fmt.Errorf("authkeys: parse %s: %w", path, parseErr)
		}
		if key.Curve != elliptic.P256() {
			return nil, false, fmt.Errorf("authkeys: %s must be a P-256 key", path)
		}
		return key, false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, false, err
	}
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, false, err
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		return nil, false, err
	}
	return key, true, nil
}

// InitSigner 在 auth-service 中调用：加载（不存在则生成）私钥，并以本机公钥作为校验密钥。
// created 表示这次是否新生成了密钥文件。
func InitSigner(keyFile string) (created bool, err error) {
	key, created, err := loadOrCreatePrivateKey(keyFile)
	if err != nil {
		return false, err
	}
	kid, err := keyID(&key.PublicKey)
	if err != nil {
		return false, err
	}
	jwks, err := json.Marshal(map[string]any{"keys": []map[string]string{jwkFor(&key.PublicKey, kid)}})
	if err != nil {
		return false, err
	}

	s := &state{signer: key, kid: kid, jwksJSON: jwks}
	s.set.replace(map[string]*ecdsa.PublicKey{kid: &key.PublicKey})
	mu.Lock()
	current = s
	mu.Unlock()
	return created, nil
}

// InitRemoteVerifier 在只需要校验的服务中调用：从 auth-service 的 JWKS 地址获取公钥。
func InitRemoteVerifier(jwksURL string) {
	s := &state{jwksURL: jwksURL, client: &http.Client{Timeout: jwksFetchTimeout}}
	mu.Lock()
	current = s
	mu.Unlock()
	if s.refresh(true) == nil {
		go s.refreshLoop(false)
		return
	}
	// auth-service 可能还没启动：持续重试直到拿到公钥，之后定时刷新
	go s.refreshLoop(true)
}

func (s *state) refreshLoop(retryUntilLoaded bool) {
	if retryUntilLoaded {
		for {
			time.Sleep(3 * time.Second)
			if s.refresh(true) == nil {
				break
			}
		}
	}
	for {
		time.Sleep(10 * time.Minute)
		_ = s.refresh(true)
	}
}

func (s *state) refresh(force bool) error {
	if s.jwksURL == "" {
		return nil
	}
	s.set.mu.Lock()
	if !force && time.Since(s.set.lastTry) < jwksMinRetryGap {
		s.set.mu.Unlock()
		return nil
	}
	s.set.lastTry = time.Now()
	s.set.mu.Unlock()

	resp, err := s.client.Get(s.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authkeys: jwks status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	keys, err := ParseJWKS(body)
	if err != nil {
		return err
	}
	s.set.replace(keys)
	return nil
}

// ParseJWKS 解析 JWKS，只接受 ES256 / P-256 公钥。
func ParseJWKS(data []byte) (map[string]*ecdsa.PublicKey, error) {
	var doc struct {
		Keys []struct {
			Kty string `json:"kty"`
			Crv string `json:"crv"`
			Kid string `json:"kid"`
			X   string `json:"x"`
			Y   string `json:"y"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	keys := make(map[string]*ecdsa.PublicKey)
	for _, k := range doc.Keys {
		if k.Kty != "EC" || k.Crv != "P-256" || k.Kid == "" {
			continue
		}
		xb, errX := base64.RawURLEncoding.DecodeString(k.X)
		yb, errY := base64.RawURLEncoding.DecodeString(k.Y)
		if errX != nil || errY != nil {
			continue
		}
		pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: new(big.Int).SetBytes(xb), Y: new(big.Int).SetBytes(yb)}
		if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
			continue
		}
		keys[k.Kid] = pub
	}
	if len(keys) == 0 {
		return nil, ErrNoKeys
	}
	return keys, nil
}

// JWKS 返回公钥集合（仅签发方可用），用于 /auth/.well-known/jwks.json。
func JWKS() ([]byte, error) {
	s := get()
	if s == nil || s.signer == nil {
		return nil, ErrNotSigner
	}
	return s.jwksJSON, nil
}

// Sign 签发 MoeLink 自家站点使用的 access token。
func Sign(userID int64, email string, ttl time.Duration) (string, error) {
	return sign(userID, email, ttl, "")
}

// SignForClient 签发发给第三方应用的受限 token：带 client_id，自家接口会拒绝它。
func SignForClient(userID int64, email string, ttl time.Duration, clientID string) (string, error) {
	return sign(userID, email, ttl, clientID)
}

func sign(userID int64, email string, ttl time.Duration, clientID string) (string, error) {
	s := get()
	if s == nil || s.signer == nil {
		return "", ErrNotSigner
	}
	jti := make([]byte, 12)
	if _, err := rand.Read(jti); err != nil {
		return "", err
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"iss":     Issuer,
		"iat":     now.Unix(),
		"exp":     now.Add(ttl).Unix(),
		"jti":     base64.RawURLEncoding.EncodeToString(jti),
	}
	if clientID != "" {
		claims["client_id"] = clientID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.kid
	return token.SignedString(s.signer)
}

// Verify 校验签名、算法、签发方与过期时间。
func Verify(tokenString string) (*Claims, error) {
	s := get()
	if s == nil {
		return nil, ErrNoKeys
	}

	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, ErrInvalidToken
		}
		if key, ok := s.set.get(kid); ok {
			return key, nil
		}
		// 密钥轮换或启动时 auth-service 尚未就绪：按需重新拉取
		_ = s.refresh(false)
		if key, ok := s.set.get(kid); ok {
			return key, nil
		}
		return nil, ErrInvalidToken
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}),
		jwt.WithIssuer(Issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	userID, okID := mc["user_id"].(float64)
	email, okEmail := mc["email"].(string)
	if !okID || !okEmail {
		return nil, ErrInvalidToken
	}
	exp, _ := mc.GetExpirationTime()
	jti, _ := mc["jti"].(string)
	clientID, _ := mc["client_id"].(string)
	claims := &Claims{UserID: int64(userID), Email: email, ID: jti, ClientID: clientID}
	if exp != nil {
		claims.ExpiresAt = exp.Time
	}
	return claims, nil
}
