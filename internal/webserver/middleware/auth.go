package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type userContextKey struct{}

func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
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

			username, ok := claims["username"].(string)
			if !ok {
				http.Error(w, "Invalid username in token", http.StatusUnauthorized)
				return
			}

			email, ok := claims["email"].(string)
			if !ok {
				http.Error(w, "Invalid email in token", http.StatusUnauthorized)
				return
			}

			// Create user claims
			userClaims := UserClaims{
				UserID:   int64(userID),
				Username: username,
				Email:    email,
			}

			// Add user claims to context
			ctx := context.WithValue(r.Context(), userContextKey{}, userClaims)

			// Continue with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(r *http.Request) (*UserClaims, bool) {
	user, ok := r.Context().Value(userContextKey{}).(UserClaims)
	return &user, ok
}

func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if token == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// claims is already a map[string]interface{}
		// Extract user information
		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			http.Error(w, "Invalid username in token", http.StatusUnauthorized)
			return
		}

		email, ok := claims["email"].(string)
		if !ok {
			http.Error(w, "Invalid email in token", http.StatusUnauthorized)
			return
		}

		// Create user claims
		userClaims := UserClaims{
			UserID:   int64(userID),
			Username: username,
			Email:    email,
		}

		// Add user claims to context
		ctx := context.WithValue(r.Context(), userContextKey{}, userClaims)

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := GetUserFromContext(r)
		if !ok || userClaims == nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "Unauthorized"})
			return
		}

		// Check if user is admin (you might want to add admin role to UserClaims)
		// For now, we'll assume admin check based on user ID or other criteria
		// This is a placeholder - implement your actual admin logic

		next.ServeHTTP(w, r)
	})
}
