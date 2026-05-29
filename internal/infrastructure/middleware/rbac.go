package middleware

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
)

// RequireRole restricts access to users holding the specified role.
func RequireRole(role entity.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles, ok := GetUserRoles(r.Context())
			if !ok {
				httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
				return
			}
			if !entity.HasRole(userRoles, role) {
				httputil.Error(w, apperror.Forbidden(apperror.CodeForbidden, "insufficient permissions", nil))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole restricts access to users holding at least one of the given roles.
func RequireAnyRole(roles ...entity.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles, ok := GetUserRoles(r.Context())
			if !ok {
				httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
				return
			}
			for _, required := range roles {
				if entity.HasRole(userRoles, required) {
					next.ServeHTTP(w, r)
					return
				}
			}
			httputil.Error(w, apperror.Forbidden(apperror.CodeForbidden, "insufficient permissions", nil))
		})
	}
}
