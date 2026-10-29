package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"

	"genshin-quiz/internal/common"
	services "genshin-quiz/internal/services/user"

	"go.uber.org/zap"
)

const maxNodeAPIRequestBody = 1 << 20

func (h *Handler) EmailCheck(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		writeNodeAPIError(w, http.StatusBadRequest, "A valid email query parameter is required")
		return
	}

	available, err := services.IsEmailAvailable(r.Context(), h.app, email)
	if err != nil {
		h.app.Logger.Error("Failed to check email availability", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeNodeAPIJSON(w, http.StatusOK, struct {
		Available bool `json:"available"`
	}{Available: available})
}

func (h *Handler) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxNodeAPIRequestBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request services.ProfileUpdateRequest
	if err := decoder.Decode(&request); err != nil {
		writeNodeAPIError(w, http.StatusBadRequest, "Invalid profile update request")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeNodeAPIError(w, http.StatusBadRequest, "Request body must contain a single JSON object")
		return
	}

	profile, err := services.UpdateProfile(r.Context(), h.app, request)
	if err != nil {
		var apiErr *common.APIError
		if errors.As(err, &apiErr) {
			writeNodeAPIError(w, apiErr.Code, apiErr.Message)
			return
		}

		h.app.Logger.Error("Failed to update user profile", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeNodeAPIJSON(w, http.StatusOK, profile)
}

func writeNodeAPIError(w http.ResponseWriter, status int, message string) {
	writeNodeAPIJSON(w, status, struct {
		Error string `json:"error"`
	}{Error: message})
}

func writeNodeAPIJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		zap.L().Error("Failed to write Node API response", zap.Error(err))
	}
}
