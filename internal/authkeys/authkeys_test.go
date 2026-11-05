package authkeys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "k.pem")
	created, err := InitSigner(path)
	if err != nil || !created {
		t.Fatalf("init: created=%v err=%v", created, err)
	}
	token, err := Sign(42, "a@b.c", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Verify(token)
	if err != nil || claims.UserID != 42 || claims.Email != "a@b.c" || claims.ID == "" {
		t.Fatalf("verify: %+v %v", claims, err)
	}

	// 重启后加载同一个密钥文件：kid 不变，旧 token 仍有效
	created, err = InitSigner(path)
	if err != nil || created {
		t.Fatalf("reload: created=%v err=%v", created, err)
	}
	if _, err := Verify(token); err != nil {
		t.Fatalf("token must survive restart: %v", err)
	}
}

func TestVerifyRejects(t *testing.T) {
	if _, err := InitSigner(filepath.Join(t.TempDir(), "k.pem")); err != nil {
		t.Fatal(err)
	}
	good, _ := Sign(1, "a@b.c", time.Minute)
	expired, _ := Sign(1, "a@b.c", -time.Minute)

	other, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	forged := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"user_id": 1, "email": "a@b.c", "iss": Issuer, "exp": time.Now().Add(time.Hour).Unix(),
	})
	forged.Header["kid"] = current.kid
	forgedStr, _ := forged.SignedString(other)

	hs := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1, "email": "a@b.c", "iss": Issuer, "exp": time.Now().Add(time.Hour).Unix(),
	})
	hs.Header["kid"] = current.kid
	hsStr, _ := hs.SignedString([]byte("secret"))

	noIss := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"user_id": 1, "email": "a@b.c", "exp": time.Now().Add(time.Hour).Unix(),
	})
	noIss.Header["kid"] = current.kid
	noIssStr, _ := noIss.SignedString(current.signer)

	for name, tok := range map[string]string{
		"expired": expired, "wrong key": forgedStr, "hs256": hsStr, "no issuer": noIssStr, "garbage": "a.b.c", "empty": "",
	} {
		if _, err := Verify(tok); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
	if _, err := Verify(good); err != nil {
		t.Fatalf("good token rejected: %v", err)
	}
}

func TestRemoteVerifier(t *testing.T) {
	if _, err := InitSigner(filepath.Join(t.TempDir(), "k.pem")); err != nil {
		t.Fatal(err)
	}
	jwks, _ := JWKS()
	token, _ := Sign(7, "x@y.z", time.Minute)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(jwks)
	}))
	defer srv.Close()

	InitRemoteVerifier(srv.URL)
	claims, err := Verify(token)
	if err != nil || claims.UserID != 7 {
		t.Fatalf("remote verify: %+v %v", claims, err)
	}
	if _, err := Sign(1, "a", time.Minute); err != ErrNotSigner {
		t.Fatalf("verifier-only process must not sign, got %v", err)
	}
}
