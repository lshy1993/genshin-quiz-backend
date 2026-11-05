package middleware

import (
	"context"
	"errors"
	"genshin-quiz/internal/authkeys"
	"genshin-quiz/internal/common"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/util"
	"net/http"
	"strings"

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
	db qrm.DB,
	requireToken bool,
	requireAdmin bool,
) func(http.Handler) http.Handler {
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
			userClaims, err := parseAndValidateToken(r, db, requireAdmin)
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
	db qrm.DB,
	requireAdmin bool,
) (*UserClaims, error) {
	// Extract Bearer token
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return nil, common.ErrInvalidTokenFormat
	}

	// 校验 ES256 签名、签发方与过期时间；公钥来自 auth-service
	claims, err := authkeys.Verify(tokenString)
	// 发给第三方应用的受限 token（带 client_id）不能访问自家接口
	if err != nil || claims.ClientID != "" {
		return nil, common.ErrInvalidToken
	}

	// 检查用户是否仍然存在于数据库中，并获取用户信息包括角色
	userInfo, err := user_repo.GetUserInfoByID(r.Context(), db, claims.UserID)
	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			return nil, common.ErrUserNotFound
		}
		return nil, err
	}

	if requireAdmin && util.IsAdmin(userInfo.UserRole) {
		return nil, common.ErrAdminAuthError
	}

	return &UserClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
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

// GenerateJWT 签发 access token，只能在持有私钥的 auth-service 进程中调用。
func GenerateJWT(userID int64, email string) (string, error) {
	return authkeys.Sign(userID, email, authkeys.AccessTokenTTL())
}

// RequiredJWTAuth 强制JWT认证中间件.
func RequiredJWTAuth(db qrm.DB) func(http.Handler) http.Handler {
	return JWTAuth(db, true, false)
}

// RequiredAdminJWTAuth 强制管理员JWT认证中间件.
func RequiredAdminJWTAuth(db qrm.DB) func(http.Handler) http.Handler {
	return JWTAuth(db, true, true)
}

// OptionalJWTAuth 可选JWT认证中间件.
func OptionalJWTAuth(db qrm.DB) func(http.Handler) http.Handler {
	return JWTAuth(db, false, false)
}

func ConditionalJWTAuth(db qrm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method

			// 检查是否是公开端点
			if isPublicEndpoint(path, method) {
				// 公开端点，使用可选认证（不强制要求token，但如果有token会解析）
				OptionalJWTAuth(db)(next).ServeHTTP(w, r)
				return
			}

			// 需要认证的端点，使用强制认证流程
			RequiredJWTAuth(db)(next).ServeHTTP(w, r)
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
		"/auth/reset-password":  {"POST"},
		"/auth/verify-email":    {"POST"},

		// 公开的只读API - 不需要认证
		"/home":        {"GET"},
		"/questions":   {"GET"},
		"/questions/*": {"GET"}, // 通配符支持 /questions/{id}
		"/exams":       {"GET"},
		"/exams/*":     {"GET"}, // 通配符支持 /exams/{id}
		"/polls":       {"GET"},
		"/polls/*":     {"GET"}, // POST/PUT需要认证
		"/users":       {"GET"},
		"/users/*":     {"GET"}, // 通配符支持 /users/{id} POST/PUT需要认证
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
