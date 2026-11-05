// Package sso 实现面向站外应用的单点登录：OAuth 2.0 授权码 + PKCE。
//
// 流程：站外应用跳转到 MoeLink 的授权页 → MoeLink 用会话 Cookie 确认用户身份并签发一次性授权码 →
// 浏览器带着授权码回到应用 → 应用用授权码 + code_verifier 换取受限 access token。
// 授权码只保存哈希、60 秒内有效、只能兑换一次；换来的 token 带 client_id，不能访问 MoeLink 自家接口。
package sso

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"genshin-quiz/internal/authkeys"
	"genshin-quiz/internal/common"

	goerrors "github.com/go-errors/errors"
)

const (
	CodeTTL        = 60 * time.Second
	ClientTokenTTL = 15 * time.Minute
)

// Client 是已登记的站外应用。
type Client struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
}

var challengePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// Clients 从 SSO_CLIENTS（JSON 数组）读取应用登记表；未配置时没有任何站外应用可以接入。
func Clients() []Client {
	raw := strings.TrimSpace(os.Getenv("SSO_CLIENTS"))
	if raw == "" {
		return nil
	}
	var clients []Client
	if err := json.Unmarshal([]byte(raw), &clients); err != nil {
		return nil
	}
	valid := clients[:0]
	for _, c := range clients {
		if c.ID != "" && len(c.RedirectURIs) > 0 {
			valid = append(valid, c)
		}
	}
	return valid
}

// FindClient 按 id 查找已登记应用。
func FindClient(id string) (*Client, bool) {
	for _, c := range Clients() {
		if c.ID == id {
			client := c
			return &client, true
		}
	}
	return nil, false
}

// ClientOrigins 返回所有已登记回调地址的 Origin，用于放行 /auth/sso/token 的跨域请求。
func ClientOrigins() []string {
	var origins []string
	for _, c := range Clients() {
		for _, uri := range c.RedirectURIs {
			if parsed, err := url.Parse(uri); err == nil && parsed.Host != "" {
				origins = append(origins, parsed.Scheme+"://"+parsed.Host)
			}
		}
	}
	return origins
}

func (c *Client) allowsRedirect(uri string) bool {
	for _, allowed := range c.RedirectURIs {
		// 必须与登记值完全一致，不做前缀 / 通配匹配
		if subtle.ConstantTimeCompare([]byte(allowed), []byte(uri)) == 1 {
			return true
		}
	}
	return false
}

func hash(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

// Authorize 为已登录用户签发授权码，返回应带回应用的回调地址。
func Authorize(
	ctx context.Context,
	db *sql.DB,
	userID int64,
	clientID, redirectURI, state, challenge, method string,
) (string, error) {
	client, ok := FindClient(clientID)
	if !ok {
		return "", common.NewBadRequestError("未知的应用")
	}
	if !client.allowsRedirect(redirectURI) {
		return "", common.NewBadRequestError("回调地址未登记")
	}
	if method != "S256" || !challengePattern.MatchString(challenge) {
		return "", common.NewBadRequestError("需要 PKCE（S256）")
	}
	if len(state) > 512 {
		return "", common.NewBadRequestError("state 过长")
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", goerrors.WrapPrefix(err, "generate code failed", 0)
	}
	code := base64.RawURLEncoding.EncodeToString(buf)
	if _, err := db.ExecContext(ctx,
		`INSERT INTO sso_auth_codes (code_hash, user_id, client_id, redirect_uri, code_challenge, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		hash(code), userID, clientID, redirectURI, challenge, time.Now().Add(CodeTTL),
	); err != nil {
		return "", goerrors.WrapPrefix(err, "store auth code failed", 0)
	}
	// 顺手清理已过期的授权码
	_, _ = db.ExecContext(ctx, `DELETE FROM sso_auth_codes WHERE expires_at < now() - interval '1 hour'`)

	target, err := url.Parse(redirectURI)
	if err != nil {
		return "", common.NewBadRequestError("回调地址无效")
	}
	query := target.Query()
	query.Set("code", code)
	if state != "" {
		query.Set("state", state)
	}
	target.RawQuery = query.Encode()
	return target.String(), nil
}

var errInvalidGrant = common.NewBadRequestError("授权码无效或已过期")

// Exchange 用授权码 + code_verifier 换取发给该应用的受限 access token。
func Exchange(
	ctx context.Context,
	db *sql.DB,
	clientID, redirectURI, code, verifier string,
) (token string, expiresIn int, err error) {
	if _, ok := FindClient(clientID); !ok {
		return "", 0, common.NewBadRequestError("未知的应用")
	}
	if len(verifier) < 43 || len(verifier) > 128 {
		return "", 0, errInvalidGrant
	}

	// 原子地“取用”授权码：同一个码只能兑换一次，并在同一条语句里校验有效期、应用与回调地址
	var userID int64
	var challenge string
	err = db.QueryRowContext(ctx,
		`UPDATE sso_auth_codes SET used_at = now()
		  WHERE code_hash = $1 AND used_at IS NULL AND expires_at > now()
		    AND client_id = $2 AND redirect_uri = $3
		RETURNING user_id, code_challenge`,
		hash(code), clientID, redirectURI,
	).Scan(&userID, &challenge)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, errInvalidGrant
	}
	if err != nil {
		return "", 0, goerrors.WrapPrefix(err, "redeem auth code failed", 0)
	}

	computed := base64.RawURLEncoding.EncodeToString(hash(verifier))
	if subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) != 1 {
		return "", 0, errInvalidGrant
	}

	var email string
	var status int16
	var deleted bool
	if err := db.QueryRowContext(ctx,
		`SELECT email, status, deleted_at IS NOT NULL FROM users WHERE id = $1`, userID,
	).Scan(&email, &status, &deleted); err != nil {
		return "", 0, goerrors.WrapPrefix(err, "load user failed", 0)
	}
	if status != 0 || deleted {
		return "", 0, errInvalidGrant
	}

	token, err = authkeys.SignForClient(userID, email, ClientTokenTTL, clientID)
	if err != nil {
		return "", 0, goerrors.WrapPrefix(err, "sign client token failed", 0)
	}
	return token, int(ClientTokenTTL.Seconds()), nil
}
