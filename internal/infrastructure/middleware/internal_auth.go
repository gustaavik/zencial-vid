package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
)

// InternalAuthHeader is the header name used to authenticate internal service-to-service calls.
const InternalAuthHeader = "X-Internal-Token"

// InternalAuth validates the X-Internal-Token header against the configured shared secret.
// Used for service-to-service callbacks (e.g. CDN → API transcode notifications).
// Rejects all requests if the secret is empty so a misconfigured deployment can't accept
// arbitrary callbacks.
func InternalAuth(sharedSecret string) func(http.Handler) http.Handler {
	expected := []byte(sharedSecret)
	configured := len(expected) > 0
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !configured {
				httputil.Unauthorized(w, apperror.CodeUnauthorized, "internal API not configured")
				return
			}
			provided := []byte(r.Header.Get(InternalAuthHeader))
			if subtle.ConstantTimeCompare(provided, expected) != 1 {
				httputil.Unauthorized(w, apperror.CodeUnauthorized, "invalid internal token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
