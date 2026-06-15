package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCountry(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		wantCode   string
		wantOK     bool
	}{
		{
			name:       "trusted proxy with CF-IPCountry",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"CF-IPCountry": "DK"},
			wantCode:   "DK",
			wantOK:     true,
		},
		{
			name:       "trusted proxy with X-Country-Code fallback",
			remoteAddr: "10.0.0.5:1234",
			headers:    map[string]string{"X-Country-Code": "us"},
			wantCode:   "US",
			wantOK:     true,
		},
		{
			name:       "CF-IPCountry takes precedence",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"CF-IPCountry": "DE", "X-Country-Code": "US"},
			wantCode:   "DE",
			wantOK:     true,
		},
		{
			name:       "untrusted public peer ignores header",
			remoteAddr: "203.0.113.10:1234",
			headers:    map[string]string{"CF-IPCountry": "DK"},
			wantOK:     false,
		},
		{
			name:       "cloudflare unknown XX maps to unknown",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"CF-IPCountry": "XX"},
			wantOK:     false,
		},
		{
			name:       "cloudflare tor T1 maps to unknown",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"CF-IPCountry": "T1"},
			wantOK:     false,
		},
		{
			name:       "invalid code rejected",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"CF-IPCountry": "DNK"},
			wantOK:     false,
		},
		{
			name:       "non-alpha code rejected",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"CF-IPCountry": "D1"},
			wantOK:     false,
		},
		{
			name:       "no header",
			remoteAddr: "127.0.0.1:1234",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotCode string
			var gotOK bool
			handler := Country(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCode, gotOK = GetCountry(r.Context())
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			handler.ServeHTTP(httptest.NewRecorder(), req)

			if gotOK != tt.wantOK {
				t.Fatalf("GetCountry ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotOK && gotCode != tt.wantCode {
				t.Errorf("GetCountry code = %q, want %q", gotCode, tt.wantCode)
			}
		})
	}
}
