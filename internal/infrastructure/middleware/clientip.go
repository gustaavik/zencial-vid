package middleware

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP rewrites r.RemoteAddr to the real client IP extracted from
// X-Real-IP or X-Forwarded-For, but only when the direct TCP connection
// originates from a private/loopback address (i.e. a trusted reverse proxy).
// Connections arriving directly from the public internet use r.RemoteAddr
// as-is, preventing header spoofing (see GHSA-3fxj-6jh8-hvhx).
func ClientIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ip := resolveClientIP(r); ip != "" {
			_, port, _ := net.SplitHostPort(r.RemoteAddr)
			if port != "" {
				r.RemoteAddr = ip + ":" + port
			} else {
				r.RemoteAddr = ip
			}
		}
		next.ServeHTTP(w, r)
	})
}

func resolveClientIP(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}
	if !isPrivateIP(remoteHost) {
		return ""
	}

	// CF-Connecting-IP is written by Cloudflare and stripped from any
	// client-supplied value, making it the most trustworthy source.
	if v := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); v != "" {
		if net.ParseIP(v) != nil {
			return v
		}
	}

	// X-Real-IP is a single authoritative value set by nginx; use it when
	// not behind Cloudflare.
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		if net.ParseIP(v) != nil {
			return v
		}
	}

	// Fall back to the leftmost entry in X-Forwarded-For, which a single-hop
	// proxy populates with the original client IP.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if v := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0]); v != "" {
			if net.ParseIP(v) != nil {
				return v
			}
		}
	}

	return ""
}

var privateRanges = func() []net.IPNet {
	cidrs := []string{
		"127.0.0.0/8",
		"::1/128",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
	}
	nets := make([]net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, _ := net.ParseCIDR(c)
		nets = append(nets, *n)
	}
	return nets
}()

func isPrivateIP(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	for _, r := range privateRanges {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}
