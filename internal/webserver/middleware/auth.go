package middleware

import (
	"context"
	user_repo "genshin-quiz/internal/repository/user"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type userContextKey struct{}

func JWTAuth(jwtSecret string, db qrm.DB) func(http.Handler) http.Handler {
	tokenAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Parse and validate token
			token, err := tokenAuth.Decode(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if token == nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Get claims from token
			claims := token.PrivateClaims()

			// Extract user information
			userID, ok := claims["user_id"].(float64)
			if !ok {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}

			email, ok := claims["email"].(string)
			if !ok {
				http.Error(w, "Invalid email in token", http.StatusUnauthorized)
				return
			}

			// 检查用户是否仍然存在于数据库中
			userExists, err := user_repo.CheckUserExists(r.Context(), db, int64(userID))
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !userExists {
				// 用户已被删除，返回特殊的强制登出响应
				writeErrorResponse(
					w,
					http.StatusUnauthorized,
					"User account no longer exists",
					"USER_DELETED",
					"Your account has been removed. Please login again.",
					true, // 强制登出
				)
				return
			} // Create user claims
			userClaims := UserClaims{
				UserID: int64(userID),
				Email:  email,
			}

			// Add user claims to context
			ctx := context.WithValue(r.Context(), userContextKey{}, userClaims)

			// Continue with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(r *http.Request) (*UserClaims, bool) {
	return GetUserFromContextOnly(r.Context())
}

func GetUserFromContextOnly(ctx context.Context) (*UserClaims, bool) {
	user, ok := ctx.Value(userContextKey{}).(UserClaims)
	return &user, ok
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := GetUserFromContext(r)
		if !ok || userClaims == nil {
			writeErrorResponse(
				w,
				http.StatusUnauthorized,
				"Unauthorized",
				"UNAUTHORIZED",
				"Admin access required",
				false, // 不需要强制登出，只是权限不足
			)
			return
		}

		// Check if user is admin (you might want to add admin role to UserClaims)
		// For now, we'll assume admin check based on user ID or other criteria
		// This is a placeholder - implement your actual admin logic

		next.ServeHTTP(w, r)
	})
}

func GenerateJWT(userID int64, email, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}

func ConditionalJWTAuth(jwtSecret string, db qrm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method

			// 检查是否是公开端点
			if isPublicEndpoint(path, method) {
				// 公开端点，不需要认证，直接通过
				next.ServeHTTP(w, r)
				return
			}

			// 需要认证的端点，使用标准认证流程
			authMiddleware := JWTAuth(jwtSecret, db)
			authMiddleware(next).ServeHTTP(w, r)
		})
	}
}

func isPublicEndpoint(path, method string) bool {
	// 定义不需要认证的路径和方法组合
	publicEndpoints := map[string][]string{
		// 认证相关 - 不需要认证
		"/auth/register":        {"POST"},
		"/auth/login":           {"POST"},
		"/auth/forgot-password": {"POST"},

		// 公开的只读API - 不需要认证
		"/questions":   {"GET"},
		"/questions/*": {"GET"}, // 通配符支持 /questions/{id}
		"/exams":       {"GET"},
		"/exams/*":     {"GET"}, // 通配符支持 /exams/{id}
		"/votes":       {"GET"},
		"/votes/*":     {"GET"}, // 只有GET操作公开，POST/PUT需要认证
	}
	// 精确匹配
	if methods, exists := publicEndpoints[path]; exists {
		for _, allowedMethod := range methods {
			if allowedMethod == method {
				return true
			}
		}
	}

	// 通配符匹配 (简单实现)
	for pattern, methods := range publicEndpoints {
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, prefix+"/") {
				for _, allowedMethod := range methods {
					if allowedMethod == method {
						return true
					}
				}
			}
		}
	}

	return false
}
