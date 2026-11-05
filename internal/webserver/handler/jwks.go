package handler

import (
	"net/http"

	"genshin-quiz/internal/authkeys"
)

// JWKS 公开 token 验证公钥，其他服务据此校验 access token，无需共享签名密钥。
func (h *Handler) JWKS(w http.ResponseWriter, _ *http.Request) {
	body, err := authkeys.JWKS()
	if err != nil {
		writeNodeAPIError(w, http.StatusServiceUnavailable, "Signing key is not available")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
