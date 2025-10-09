package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"

	"genshin-quiz/config"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

func Handler(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 手动捕获错误到Sentry (如果已初始化)
					sentry.WithScope(func(scope *sentry.Scope) {
						scope.SetRequest(r)
						scope.SetLevel(sentry.LevelError)
						scope.SetTag("error_type", "panic")
						scope.SetContext("request", map[string]interface{}{
							"method":  r.Method,
							"url":     r.URL.String(),
							"headers": r.Header,
						})
						sentry.CaptureException(fmt.Errorf("panic recovered: %v", err))
					})

					logger.Error("Panic recovered",
						zap.String("method", r.Method),
						zap.String("url", r.URL.String()),
						zap.Any("error", err),
						zap.String("request_id", r.Header.Get("X-Request-ID")),
					)

					writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", "", "")
				}
			}()

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, message, code, details string) {
	writeErrorResponse(w, statusCode, message, code, details)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message, code, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		// If JSON encoding fails, write a simple error message
		// We intentionally ignore the error from Write as there's nothing more we can do
		w.Write([]byte(`{"error":"Internal server error"}`)) //nolint:errcheck // fallback error writing
	}
}

func HandleBadRequestError(app *config.App) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		app.Logger.Error("Bad request error",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Error(err),
			zap.String("request_id", r.Header.Get("X-Request-ID")),
		)

		writeErrorResponse(w, http.StatusBadRequest, "Bad request", "INVALID_REQUEST", err.Error())
	}
}

func HandleResponseErrorWithLog(app *config.App) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		// 手动捕获错误到Sentry (如果已初始化)
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetRequest(r)
			scope.SetLevel(sentry.LevelError)
			scope.SetTag("error_type", "response_error")
			scope.SetContext("request", map[string]interface{}{
				"method":  r.Method,
				"url":     r.URL.String(),
				"headers": r.Header,
			})
			sentry.CaptureException(err)
		})

		app.Logger.Error("Response error",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Error(err),
			zap.String("request_id", r.Header.Get("X-Request-ID")),
		)

		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR", err.Error())
	}
}
