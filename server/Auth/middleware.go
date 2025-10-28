package Auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware validates JWT token and adds user to context
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "missing authorization header",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeJSON(w, http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "invalid authorization header format",
			})
			return
		}

		tokenString := parts[1]

		// Check blacklist
		if IsTokenBlacklisted(tokenString) {
			writeJSON(w, http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "token has been revoked",
			})
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "invalid token",
			})
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	return claims, ok
}
