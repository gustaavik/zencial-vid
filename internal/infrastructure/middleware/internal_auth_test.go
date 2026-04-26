package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInternalAuth(t *testing.T) {
	const secret = "s3cr3t"
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	t.Run("missing header rejects with 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/x", http.NoBody)
		rec := httptest.NewRecorder()
		InternalAuth(secret)(okHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("wrong token rejects with 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/x", http.NoBody)
		req.Header.Set(InternalAuthHeader, "nope")
		rec := httptest.NewRecorder()
		InternalAuth(secret)(okHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("matching token passes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/x", http.NoBody)
		req.Header.Set(InternalAuthHeader, secret)
		rec := httptest.NewRecorder()
		InternalAuth(secret)(okHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("empty secret rejects everything", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/x", http.NoBody)
		req.Header.Set(InternalAuthHeader, "")
		rec := httptest.NewRecorder()
		InternalAuth("")(okHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "unconfigured server must reject all callbacks")
	})
}
