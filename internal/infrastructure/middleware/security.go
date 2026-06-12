package middleware

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// Security sets HSTS and X-Content-Type-Options on every response.
// When APP_ENV is not "development" and the reverse proxy signals that the
// original request arrived over plain HTTP (X-Forwarded-Proto: http), it
// issues a permanent redirect to the HTTPS equivalent of the URL.
func Security(cfg config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			w.Header().Set("X-Content-Type-Options", "nosniff")

			if cfg.AppEnv != "development" && r.Header.Get("X-Forwarded-Proto") == "http" {
				http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
