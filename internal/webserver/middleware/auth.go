package middleware

import (
	"context"
	"errors"
	"fmt"
	"genshin-quiz/internal/common"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/util"
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

func JWTAuth(
	jwtSecret string,
	db qrm.DB,
	requireToken bool,
	requireAdmin bool,
) func(http.Handler) http.Handler {
	tokenAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否有token
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				// 强制token，不存在则报错
				if requireToken {
					http.Error(w, "Missing authorization header", http.StatusUnauthorized)
					return
				}
				// 可选认证，直接继续
				next.ServeHTTP(w, r)
				return
			}

			// 提供了token，则需要验证
			userClaims, err := parseAndValidateToken(r, tokenAuth, db, requireAdmin)
			if err != nil {
				// token无效 - 不管是强制还是可选认证都应该报错
				handleAuthError(w, err)
				return
			}

			// token有效，添加用户信息到context
			ctx := context.WithValue(r.Context(), userContextKey{}, *userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseAndValidateToken(
	r *http.Request,
	tokenAuth *jwtauth.JWTAuth,
	db qrm.DB,
	requireAdmin bool,
) (*UserClaims, error) {
	// Extract Bearer token
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return nil, common.ErrInvalidTokenFormat
	}

	// Parse and validate token (包括过期检查)
	token, err := tokenAuth.Decode(tokenString)
	if err != nil || token == nil {
		return nil, common.ErrInvalidToken
	}

	// Get claims from token
	claims := token.PrivateClaims()

	// Extract user information
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, common.ErrInvalidToken
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, common.ErrInvalidToken
	}

	// 检查用户是否仍然存在于数据库中，并获取用户信息包括角色
	userInfo, err := user_repo.GetUserInfoByID(r.Context(), db, int64(userID))
	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			return nil, common.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}

	if requireAdmin && util.IsAdmin(*userInfo.UserRole) {
		return nil, common.ErrAdminAuthError
	}

	return &UserClaims{
		UserID: int64(userID),
		Email:  email,
	}, nil
}

func handleAuthError(w http.ResponseWriter, err error) {
	if errors.Is(err, common.ErrUserNotFound) {
		writeErrorResponse(
			w,
			http.StatusUnauthorized,
			"User account no longer exists",
			"USER_DELETED",
			"Your account has been removed. Please login again.",
			true, // 强制登出
		)
		return
	}

	if errors.Is(err, common.ErrDatabaseError) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 其他认证错误（token格式错误、过期等）
	if errors.Is(err, common.ErrInvalidToken) || errors.Is(err, common.ErrInvalidTokenFormat) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 默认处理
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

func GetUserFromContext(r *http.Request) (*UserClaims, bool) {
	return GetUserFromContextOnly(r.Context())
}

func GetUserFromContextOnly(ctx context.Context) (*UserClaims, bool) {
	user, ok := ctx.Value(userContextKey{}).(UserClaims)
	return &user, ok
}

func GenerateJWT(userID int64, email, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}

// RequiredJWTAuth 强制JWT认证中间件.
func RequiredJWTAuth(jwtSecret string, db qrm.DB) func(http.Handler) http.Handler {
	return JWTAuth(jwtSecret, db, true, false)
}

// RequiredAdminJWTAuth 强制管理员JWT认证中间件.
func RequiredAdminJWTAuth(jwtSecret string, db qrm.DB) func(http.Handler) http.Handler {
	return JWTAuth(jwtSecret, db, true, true)
}

// OptionalJWTAuth 可选JWT认证中间件.
func OptionalJWTAuth(jwtSecret string, db qrm.DB) func(http.Handler) http.Handler {
	return JWTAuth(jwtSecret, db, false, false)
}

func ConditionalJWTAuth(jwtSecret string, db qrm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method

			// 检查是否是公开端点
			if isPublicEndpoint(path, method) {
				// 公开端点，使用可选认证（不强制要求token，但如果有token会解析）
				OptionalJWTAuth(jwtSecret, db)(next).ServeHTTP(w, r)
				return
			}

			// 需要认证的端点，使用强制认证流程
			RequiredJWTAuth(jwtSecret, db)(next).ServeHTTP(w, r)
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
