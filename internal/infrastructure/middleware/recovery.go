package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/zenfulcode/zencial/internal/pkg/httputil"
)

// Recovery catches panics and returns a 500 error.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						"error", rec,
						"stack", string(debug.Stack()),
						"request_id", GetRequestID(r.Context()),
						"path", r.URL.Path,
					)
					httputil.InternalError(w)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
