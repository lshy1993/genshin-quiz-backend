package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"genshin-quiz/internal/common"
	"genshin-quiz/internal/enum"
	services "genshin-quiz/internal/services/user"
	mw "genshin-quiz/internal/webserver/middleware"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (h *Handler) writeIdentityError(w http.ResponseWriter, message string, err error) {
	var apiErr *common.APIError
	if errors.As(err, &apiErr) {
		writeNodeAPIError(w, apiErr.Code, apiErr.Message)
		return
	}
	h.app.Logger.Error(message, zap.Error(err))
	writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
}

func (h *Handler) ListIdentities(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetUserFromContext(r)
	if !ok {
		writeNodeAPIError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	result, err := services.ListIdentities(r.Context(), h.app, claims.UserID)
	if err != nil {
		h.writeIdentityError(w, "Failed to list identities", err)
		return
	}
	writeNodeAPIJSON(w, http.StatusOK, result)
}

func (h *Handler) LinkIdentity(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetUserFromContext(r)
	if !ok {
		writeNodeAPIError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxIdentityBodySize)
	var request struct {
		LinkToken string `json:"link_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.LinkToken == "" {
		writeNodeAPIError(w, http.StatusBadRequest, "link_token is required")
		return
	}
	link, err := verifyOAuthLinkToken(h.app.Config.JWTSecret, request.LinkToken)
	if err != nil {
		writeNodeAPIError(w, http.StatusBadRequest, "绑定凭证无效或已过期，请重新绑定")
		return
	}
	identity := services.OAuthIdentity{
		Provider: enum.LoginProvider(link.Provider), Subject: link.Subject, Email: link.Email,
		EmailVerified: true, Name: link.Name, AvatarURL: link.AvatarURL,
	}
	if err := services.LinkIdentity(r.Context(), h.app, claims.UserID, identity); err != nil {
		h.writeIdentityError(w, "Failed to link identity", err)
		return
	}
	writeNodeAPIJSON(w, http.StatusOK, struct {
		Success  bool   `json:"success"`
		Provider string `json:"provider"`
	}{Success: true, Provider: link.Provider})
}

func (h *Handler) UnlinkIdentity(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetUserFromContext(r)
	if !ok {
		writeNodeAPIError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err := services.UnlinkIdentity(r.Context(), h.app, claims.UserID, chi.URLParam(r, "provider")); err != nil {
		h.writeIdentityError(w, "Failed to unlink identity", err)
		return
	}
	writeNodeAPIJSON(w, http.StatusOK, struct {
		Success bool `json:"success"`
	}{Success: true})
}

func (h *Handler) SetPassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetUserFromContext(r)
	if !ok {
		writeNodeAPIError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxIdentityBodySize)
	var request struct {
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeNodeAPIError(w, http.StatusBadRequest, "new_password is required")
		return
	}
	if err := services.SetPassword(r.Context(), h.app, claims.UserID, request.NewPassword); err != nil {
		h.writeIdentityError(w, "Failed to set password", err)
		return
	}
	writeNodeAPIJSON(w, http.StatusOK, map[string]bool{"success": true})
}
