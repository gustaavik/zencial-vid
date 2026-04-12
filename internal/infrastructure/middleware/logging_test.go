package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturingHandler captures the last log record for assertions.
type capturingHandler struct {
	last *slog.Record
}

func (h *capturingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	h.last = &r
	return nil
}
func (h *capturingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *capturingHandler) WithGroup(string) slog.Handler      { return h }

func TestLogger_LevelSelection(t *testing.T) {
	tests := []struct {
		name        string
		setupRouter func(r *chi.Mux)
		method      string
		path        string
		wantLevel   slog.Level
	}{
		{
			name: "matched route 200 logs INFO",
			setupRouter: func(r *chi.Mux) {
				r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			method:    http.MethodGet,
			path:      "/health",
			wantLevel: slog.LevelInfo,
		},
		{
			name: "matched route returning 404 logs WARN",
			setupRouter: func(r *chi.Mux) {
				r.Get("/api/v1/videos/{id}", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				})
			},
			method:    http.MethodGet,
			path:      "/api/v1/videos/nonexistent",
			wantLevel: slog.LevelWarn,
		},
		{
			name: "unmatched route /p.php logs DEBUG",
			setupRouter: func(r *chi.Mux) {
				r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			method:    http.MethodGet,
			path:      "/p.php",
			wantLevel: slog.LevelDebug,
		},
		{
			name: "unmatched route /admin/phpinfo.php logs DEBUG",
			setupRouter: func(r *chi.Mux) {
				r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			method:    http.MethodGet,
			path:      "/admin/phpinfo.php",
			wantLevel: slog.LevelDebug,
		},
		{
			name: "matched route 500 logs ERROR",
			setupRouter: func(r *chi.Mux) {
				r.Get("/boom", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				})
			},
			method:    http.MethodGet,
			path:      "/boom",
			wantLevel: slog.LevelError,
		},
		{
			name: "405 method not allowed logs WARN",
			setupRouter: func(r *chi.Mux) {
				r.Get("/resource", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			method:    http.MethodPost,
			path:      "/resource",
			wantLevel: slog.LevelWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &capturingHandler{}
			log := slog.New(h)

			r := chi.NewRouter()
			r.Use(Logger(log))
			tt.setupRouter(r)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.NotNil(t, h.last, "expected a log record")
			assert.Equal(t, tt.wantLevel, h.last.Level,
				"unexpected log level for %s %s (status %d)", tt.method, tt.path, w.Code)
		})
	}
}
