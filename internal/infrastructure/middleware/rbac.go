package middleware

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
)

// RequireRole restricts access to users with the specified role.
func RequireRole(role entity.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRole(r.Context())
			if !ok {
				httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
				return
			}
			if userRole != role {
				httputil.Error(w, apperror.Forbidden(apperror.CodeForbidden, "insufficient permissions", nil))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
