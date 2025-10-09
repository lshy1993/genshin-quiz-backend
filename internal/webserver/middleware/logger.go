package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// HTTP状态码颜色映射
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "\033[32m" // 绿色 - 成功
	case status >= 300 && status < 400:
		return "\033[33m" // 黄色 - 重定向
	case status >= 400 && status < 500:
		return "\033[35m" // 紫色 - 客户端错误
	case status >= 500:
		return "\033[31m" // 红色 - 服务器错误
	default:
		return "\033[0m" // 默认颜色
	}
}

// HTTP方法颜色映射
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[34m" // 蓝色
	case "POST":
		return "\033[32m" // 绿色
	case "PUT":
		return "\033[33m" // 黄色
	case "DELETE":
		return "\033[31m" // 红色
	case "PATCH":
		return "\033[35m" // 紫色
	case "HEAD":
		return "\033[36m" // 青色
	default:
		return "\033[0m" // 默认颜色
	}
}

const resetColor = "\033[0m"

// formatHTTPRequest 创建带颜色的HTTP请求日志消息
func formatHTTPRequest(method, url string, status int, duration time.Duration) string {
	methodColor := getMethodColor(method)
	statusColor := getStatusColor(status)

	return fmt.Sprintf("%s%s%s %s -> %s%d%s (%v)",
		methodColor, method, resetColor,
		url,
		statusColor, status, resetColor,
		duration,
	)
}

func Logger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a wrapped response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				duration := time.Since(start)
				status := ww.Status()

				// 使用带颜色的自定义格式
				message := formatHTTPRequest(r.Method, r.URL.String(), status, duration)

				logger.Info(message,
					zap.String("method", r.Method),
					zap.String("url", r.URL.String()),
					zap.Int("status", status),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("user_agent", r.UserAgent()),
					zap.Int("bytes", ww.BytesWritten()),
					zap.Duration("duration", duration),
					zap.String("request_id", middleware.GetReqID(r.Context())),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

func ErrorLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("method", r.Method),
						zap.String("url", r.URL.String()),
						zap.String("request_id", middleware.GetReqID(r.Context())),
					)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
