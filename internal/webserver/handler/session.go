package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"genshin-quiz/internal/authkeys"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/services/session"
	"genshin-quiz/internal/services/sso"

	"go.uber.org/zap"
)

// IsTrustedOrigin 判断 Origin 是否属于受信任的站点（与 OAuth 回跳白名单一致）。
func IsTrustedOrigin(origin string) bool {
	for _, allowed := range allowedOAuthOrigins() {
		if origin == allowed {
			return true
		}
	}
	return false
}

// IsSSOClientOrigin 判断 Origin 是否属于已登记的站外应用（仅用于放行 /auth/sso/token 的跨域读取）。
func IsSSOClientOrigin(origin string) bool {
	for _, allowed := range sso.ClientOrigins() {
		if origin == allowed {
			return true
		}
	}
	return false
}

// originAllowed 保护依赖 Cookie 的接口：带 Origin 的跨站请求必须来自受信任站点。
func originAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if parsed, err := url.Parse(origin); err != nil || parsed.Host == "" {
		return false
	}
	return IsTrustedOrigin(origin)
}

func noStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

// SessionRefresh 用会话 Cookie 换取短期 access token。
func (h *Handler) SessionRefresh(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if !originAllowed(r) {
		writeNodeAPIError(w, http.StatusForbidden, "Origin not allowed")
		return
	}
	_, userID, err := session.Active(r.Context(), h.app.DB, r)
	if err != nil {
		if errors.Is(err, session.ErrNoSession) {
			writeNodeAPIError(w, http.StatusUnauthorized, "No active session")
			return
		}
		h.app.Logger.Error("Failed to check session", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	user, err := user_repo.GetUserInfoByID(r.Context(), h.app.DB, userID)
	if err != nil || user.Status != 0 || user.DeletedAt != nil {
		_ = session.Revoke(r.Context(), h.app.DB, r)
		session.Clear(w)
		writeNodeAPIError(w, http.StatusUnauthorized, "No active session")
		return
	}

	ttl := authkeys.AccessTokenTTL()
	token, err := authkeys.Sign(user.ID, user.Email, ttl)
	if err != nil {
		h.app.Logger.Error("Failed to sign access token", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeNodeAPIJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_in": int(ttl.Seconds()),
	})
}

// Logout 撤销当前会话并清除 Cookie。
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if !originAllowed(r) {
		writeNodeAPIError(w, http.StatusForbidden, "Origin not allowed")
		return
	}
	if err := session.Revoke(r.Context(), h.app.DB, r); err != nil {
		h.app.Logger.Error("Failed to revoke session", zap.Error(err))
	}
	session.Clear(w)
	w.WriteHeader(http.StatusNoContent)
}

// LogoutAll 撤销该用户在所有设备上的全部会话。
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if !originAllowed(r) {
		writeNodeAPIError(w, http.StatusForbidden, "Origin not allowed")
		return
	}
	_, userID, err := session.Active(r.Context(), h.app.DB, r)
	if err != nil {
		writeNodeAPIError(w, http.StatusUnauthorized, "No active session")
		return
	}
	if err := session.RevokeAll(r.Context(), h.app.DB, userID); err != nil {
		h.app.Logger.Error("Failed to revoke all sessions", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	session.Clear(w)
	w.WriteHeader(http.StatusNoContent)
}

// SSOClientInfo 返回已登记应用的展示信息，供 MoeLink 授权页显示“正在登录到 xxx”。
func (h *Handler) SSOClientInfo(w http.ResponseWriter, r *http.Request) {
	client, ok := sso.FindClient(r.URL.Query().Get("client_id"))
	if !ok {
		writeNodeAPIError(w, http.StatusNotFound, "未知的应用")
		return
	}
	name := client.Name
	if name == "" {
		name = client.ID
	}
	writeNodeAPIJSON(w, http.StatusOK, map[string]string{"id": client.ID, "name": name})
}

// SSOAuthorize 在 MoeLink 授权页确认用户（会话 Cookie）后，签发站外应用的授权码。
func (h *Handler) SSOAuthorize(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if !originAllowed(r) {
		writeNodeAPIError(w, http.StatusForbidden, "Origin not allowed")
		return
	}
	_, userID, err := session.Active(r.Context(), h.app.DB, r)
	if err != nil {
		writeNodeAPIError(w, http.StatusUnauthorized, "No active session")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxIdentityBodySize)
	var request struct {
		ClientID            string `json:"client_id"`
		RedirectURI         string `json:"redirect_uri"`
		State               string `json:"state"`
		CodeChallenge       string `json:"code_challenge"`
		CodeChallengeMethod string `json:"code_challenge_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeNodeAPIError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	target, err := sso.Authorize(r.Context(), h.app.DB, userID,
		request.ClientID, request.RedirectURI, request.State, request.CodeChallenge, request.CodeChallengeMethod)
	if err != nil {
		h.writeIdentityError(w, "Failed to authorize SSO client", err)
		return
	}
	writeNodeAPIJSON(w, http.StatusOK, map[string]string{"redirect_to": target})
}

// SSOToken 用授权码 + code_verifier 换取发给站外应用的受限 access token（公开客户端，不使用 Cookie）。
func (h *Handler) SSOToken(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	r.Body = http.MaxBytesReader(w, r.Body, maxIdentityBodySize)
	var request struct {
		ClientID     string `json:"client_id"`
		RedirectURI  string `json:"redirect_uri"`
		Code         string `json:"code"`
		CodeVerifier string `json:"code_verifier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeNodeAPIError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	token, expiresIn, err := sso.Exchange(r.Context(), h.app.DB,
		request.ClientID, request.RedirectURI, request.Code, request.CodeVerifier)
	if err != nil {
		h.writeIdentityError(w, "Failed to exchange SSO code", err)
		return
	}
	writeNodeAPIJSON(w, http.StatusOK, map[string]any{
		"token_type": "Bearer", "token": token, "expires_in": expiresIn,
	})
}
