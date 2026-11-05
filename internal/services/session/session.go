// Package session 管理登录会话：浏览器持有不透明的 HttpOnly Cookie，
// 服务端据此签发短期 access token，并支持随时撤销。
package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"genshin-quiz/internal/common"
	"genshin-quiz/internal/webserver/middleware"

	goerrors "github.com/go-errors/errors"
)

const (
	CookieName = "moelink_session"
	// SessionTTL 为滑动有效期：每次换取 access token 都会顺延。
	SessionTTL = 30 * 24 * time.Hour
)

var ErrNoSession = errors.New("session: missing or invalid session")

// Execer 由 *sql.DB、*sql.Tx 以及 jet 的 qrm.DB 共同满足。
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Querier 用于需要读取一行结果的操作。
type Querier interface {
	Execer
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func hashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func cookieSecure() bool {
	return strings.HasPrefix(os.Getenv("OAUTH_CALLBACK_BASE_URL"), "https://")
}

// cookieDomain 在线上设置为 .moelink.site，使所有子站共享会话；本地留空（localhost 的 Cookie 不区分端口）。
func cookieDomain() string {
	return strings.TrimSpace(os.Getenv("SESSION_COOKIE_DOMAIN"))
}

func writeCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		Domain:   cookieDomain(),
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
}

// Start 为用户创建新会话；如果 ctx 中带有响应写入器，同时下发会话 Cookie。
func Start(ctx context.Context, db Execer, userID int64) error {
	raw, err := newToken()
	if err != nil {
		return goerrors.WrapPrefix(err, "generate session token failed", 0)
	}
	ip, _ := ctx.Value(middleware.RealIPKey).(string)
	ua, _ := ctx.Value(middleware.UserAgentKey).(string)
	if len(ua) > 300 {
		ua = ua[:300]
	}
	// 顺手清掉该用户已过期 / 已撤销的旧会话，避免表无限增长
	_, _ = db.ExecContext(ctx,
		`DELETE FROM user_sessions WHERE user_id = $1 AND (expires_at < now() OR revoked_at IS NOT NULL)`, userID)
	if _, err := db.ExecContext(ctx,
		`INSERT INTO user_sessions (user_id, token_hash, expires_at, ip, user_agent) VALUES ($1, $2, $3, $4, $5)`,
		userID, hashToken(raw), time.Now().Add(SessionTTL), ip, ua,
	); err != nil {
		return goerrors.WrapPrefix(err, "create session failed", 0)
	}
	if w, ok := middleware.ResponseWriterFromContext(ctx); ok {
		writeCookie(w, raw, int(SessionTTL.Seconds()))
	}
	return nil
}

// Active 校验请求里的会话 Cookie，返回会话 id 与用户 id，并顺延有效期。
func Active(ctx context.Context, db Querier, r *http.Request) (sessionID, userID int64, err error) {
	cookie, cookieErr := r.Cookie(CookieName)
	if cookieErr != nil || cookie.Value == "" {
		return 0, 0, ErrNoSession
	}
	err = db.QueryRowContext(ctx,
		`UPDATE user_sessions
		   SET last_used_at = now(), expires_at = $2
		 WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
		 RETURNING id, user_id`,
		hashToken(cookie.Value), time.Now().Add(SessionTTL),
	).Scan(&sessionID, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, ErrNoSession
	}
	if err != nil {
		return 0, 0, goerrors.WrapPrefix(err, "lookup session failed", 0)
	}
	return sessionID, userID, nil
}

// Revoke 撤销当前请求携带的会话。
func Revoke(ctx context.Context, db Execer, r *http.Request) error {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	_, err = db.ExecContext(ctx,
		`UPDATE user_sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		hashToken(cookie.Value))
	return err
}

// RevokeAll 撤销用户的全部会话（所有设备登出）。
func RevokeAll(ctx context.Context, db Execer, userID int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE user_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

// Clear 让浏览器删除会话 Cookie。
func Clear(w http.ResponseWriter) {
	writeCookie(w, "", -1)
}

// Require 在会话无效时返回 401 风格的业务错误。
func Require(err error) error {
	if errors.Is(err, ErrNoSession) {
		return common.ErrInvalidToken
	}
	return err
}
