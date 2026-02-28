package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
)

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

// Authenticate validates the JWT access token and stores claims in context.
func Authenticate(tokenService auth.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httputil.Unauthorized(w, "UNAUTHORIZED", "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				httputil.Unauthorized(w, "UNAUTHORIZED", "invalid authorization format")
				return
			}

			claims, err := tokenService.ValidateAccessToken(parts[1])
			if err != nil {
				httputil.Unauthorized(w, "INVALID_TOKEN", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, userRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves the authenticated user's ID from context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

// GetUserRole retrieves the authenticated user's role from context.
func GetUserRole(ctx context.Context) (entity.UserRole, bool) {
	role, ok := ctx.Value(userRoleKey).(entity.UserRole)
	return role, ok
}
