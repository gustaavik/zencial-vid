package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

const countryKey contextKey = "country_code"

// Country stores the client's ISO 3166-1 alpha-2 country code in the request
// context, read from CF-IPCountry (Cloudflare) or X-Country-Code (other
// trusted proxies). Headers are only honored when the direct TCP connection
// originates from a private/loopback address, mirroring ClientIP's trust
// model. Because ClientIP rewrites r.RemoteAddr to the public client address,
// this middleware MUST be registered before ClientIP.
func Country(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if code := resolveCountry(r); code != "" {
			r = r.WithContext(context.WithValue(r.Context(), countryKey, code))
		}
		next.ServeHTTP(w, r)
	})
}

// GetCountry retrieves the client's country code from context. The second
// return value is false when no trusted country information was available.
func GetCountry(ctx context.Context) (string, bool) {
	code, ok := ctx.Value(countryKey).(string)
	return code, ok
}

func resolveCountry(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}
	if !isPrivateIP(remoteHost) {
		return ""
	}

	for _, header := range []string{"CF-IPCountry", "X-Country-Code"} {
		if code := normalizeCountryCode(r.Header.Get(header)); code != "" {
			return code
		}
	}
	return ""
}

// normalizeCountryCode validates and upper-cases a 2-letter country code.
// Cloudflare's special values XX (unknown) and T1 (Tor) map to unknown.
func normalizeCountryCode(raw string) string {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if len(code) != 2 {
		return ""
	}
	if code == "XX" || code == "T1" {
		return ""
	}
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			return ""
		}
	}
	return code
}
