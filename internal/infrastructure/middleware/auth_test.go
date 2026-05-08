package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// --- minimal mocks scoped to this test ---

type stubSessionRepo struct {
	getByTokenHashFn func(ctx context.Context, hash string) (*entity.Session, error)
	updateActivityFn func(ctx context.Context, id uuid.UUID, last, idle time.Time) error
	updateCount      atomic.Int32
}

func (s *stubSessionRepo) Create(context.Context, *entity.Session) error { return nil }
func (s *stubSessionRepo) GetByTokenHash(ctx context.Context, hash string) (*entity.Session, error) {
	return s.getByTokenHashFn(ctx, hash)
}
func (s *stubSessionRepo) GetByID(context.Context, uuid.UUID) (*entity.Session, error) {
	return nil, nil
}
func (s *stubSessionRepo) ListByUserID(context.Context, uuid.UUID, *filter.FilterSet) ([]entity.Session, int64, error) {
	return nil, 0, nil
}
func (s *stubSessionRepo) UpdateActivity(ctx context.Context, id uuid.UUID, last, idle time.Time) error {
	s.updateCount.Add(1)
	if s.updateActivityFn != nil {
		return s.updateActivityFn(ctx, id, last, idle)
	}
	return nil
}
func (s *stubSessionRepo) Revoke(context.Context, uuid.UUID, time.Time) error            { return nil }
func (s *stubSessionRepo) RevokeAllForUser(context.Context, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (s *stubSessionRepo) RevokeOthersForUser(context.Context, uuid.UUID, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (s *stubSessionRepo) DeleteExpired(context.Context, time.Time) (int64, error) { return 0, nil }

type stubUserRepo struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

func (u *stubUserRepo) Create(context.Context, *entity.User) error { return nil }
func (u *stubUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if u.getByIDFn != nil {
		return u.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (u *stubUserRepo) GetByEmail(context.Context, valueobject.Email) (*entity.User, error) {
	return nil, nil
}
func (u *stubUserRepo) Update(context.Context, *entity.User) error { return nil }
func (u *stubUserRepo) Delete(context.Context, uuid.UUID) error    { return nil }
func (u *stubUserRepo) ExistsByEmail(context.Context, valueobject.Email) (bool, error) {
	return false, nil
}
func (u *stubUserRepo) List(context.Context, *filter.FilterSet) ([]entity.User, int64, error) {
	return nil, 0, nil
}
func (u *stubUserRepo) UpdateStatus(context.Context, uuid.UUID, entity.UserStatus) error {
	return nil
}

type stubTokens struct{}

func (stubTokens) Generate() (string, string, error) { return "tok", "hash", nil }
func (stubTokens) Hash(t string) string               { return "hash:" + t }

// --- helpers ---

func testCfg() config.SessionConfig {
	return config.SessionConfig{
		IdleTimeout:     time.Hour,
		AbsoluteTimeout: 24 * time.Hour,
		SlideDebounce:   5 * time.Minute,
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newActiveTestUser(id uuid.UUID) *entity.User {
	u := entity.NewUser(valueobject.EmailFromTrusted("a@b.com"), valueobject.NewHashedPassword("h"))
	u.ID = id
	return u
}

func makeSession(now time.Time) *entity.Session {
	s := entity.NewSession(uuid.New(), "hash:tok", "dev", "ua", "ip", now, time.Hour, 24*time.Hour)
	return s
}

func decodeErrorCode(t *testing.T, body []byte) string {
	t.Helper()
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp.Error.Code
}

// --- tests ---

func TestAuthenticate(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	clk := clock.MockClock{FixedTime: now}
	cfg := testCfg()
	log := discardLogger()

	t.Run("missing header returns 401 unauthorized", func(t *testing.T) {
		auth := NewSessionAuthenticator(&stubSessionRepo{}, &stubUserRepo{}, stubTokens{}, clk, cfg, log)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		auth.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			t.Fatal("handler should not be invoked")
		})).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, apperror.CodeUnauthorized, decodeErrorCode(t, rr.Body.Bytes()))
	})

	t.Run("malformed header returns 401", func(t *testing.T) {
		auth := NewSessionAuthenticator(&stubSessionRepo{}, &stubUserRepo{}, stubTokens{}, clk, cfg, log)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "not-bearer xyz")
		rr := httptest.NewRecorder()
		auth.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("unknown token returns 401 INVALID_TOKEN", func(t *testing.T) {
		repo := &stubSessionRepo{getByTokenHashFn: func(context.Context, string) (*entity.Session, error) {
			return nil, nil
		}}
		auth := NewSessionAuthenticator(repo, &stubUserRepo{}, stubTokens{}, clk, cfg, log)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rr := httptest.NewRecorder()
		auth.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, apperror.CodeInvalidToken, decodeErrorCode(t, rr.Body.Bytes()))
	})

	t.Run("revoked session returns 401 SESSION_REVOKED", func(t *testing.T) {
		s := makeSession(now)
		rev := s.Revoke(now)
		repo := &stubSessionRepo{getByTokenHashFn: func(context.Context, string) (*entity.Session, error) {
			return &rev, nil
		}}
		auth := NewSessionAuthenticator(repo, &stubUserRepo{}, stubTokens{}, clk, cfg, log)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rr := httptest.NewRecorder()
		auth.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, apperror.CodeSessionRevoked, decodeErrorCode(t, rr.Body.Bytes()))
	})

	t.Run("expired session returns 401 SESSION_EXPIRED", func(t *testing.T) {
		s := makeSession(now.Add(-2 * time.Hour))
		repo := &stubSessionRepo{getByTokenHashFn: func(context.Context, string) (*entity.Session, error) {
			return s, nil
		}}
		auth := NewSessionAuthenticator(repo, &stubUserRepo{}, stubTokens{}, clk, cfg, log)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rr := httptest.NewRecorder()
		auth.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, apperror.CodeSessionExpired, decodeErrorCode(t, rr.Body.Bytes()))
	})

	t.Run("suspended user returns 401 unauthorized", func(t *testing.T) {
		s := makeSession(now)
		userID := s.UserID
		u := newActiveTestUser(userID)
		u.Suspend()
		repo := &stubSessionRepo{getByTokenHashFn: func(context.Context, string) (*entity.Session, error) {
			return s, nil
		}}
		userRepo := &stubUserRepo{getByIDFn: func(context.Context, uuid.UUID) (*entity.User, error) {
			return u, nil
		}}
		auth := NewSessionAuthenticator(repo, userRepo, stubTokens{}, clk, cfg, log)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rr := httptest.NewRecorder()
		auth.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("happy path within debounce does not slide", func(t *testing.T) {
		s := makeSession(now)
		userID := s.UserID
		u := newActiveTestUser(userID)
		repo := &stubSessionRepo{getByTokenHashFn: func(context.Context, string) (*entity.Session, error) {
			return s, nil
		}}
		userRepo := &stubUserRepo{getByIDFn: func(context.Context, uuid.UUID) (*entity.User, error) {
			return u, nil
		}}
		auth := NewSessionAuthenticator(repo, userRepo, stubTokens{}, clk, cfg, log)

		var capturedUserID uuid.UUID
		var capturedSessionID uuid.UUID
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rr := httptest.NewRecorder()
		auth.Authenticate(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			capturedUserID, _ = GetUserID(r.Context())
			capturedSessionID, _ = GetSessionID(r.Context())
		})).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, userID, capturedUserID)
		assert.Equal(t, s.ID, capturedSessionID)
		assert.EqualValues(t, 0, repo.updateCount.Load(),
			"slide debounce should suppress UpdateActivity for fresh session")
	})

	t.Run("happy path past debounce slides activity", func(t *testing.T) {
		// Build a session whose LastActivityAt is comfortably past the debounce.
		s := makeSession(now.Add(-30 * time.Minute))
		s.LastActivityAt = now.Add(-30 * time.Minute)
		s.IdleExpiresAt = now.Add(30 * time.Minute) // still valid
		s.AbsoluteExpiresAt = now.Add(23 * time.Hour)

		u := newActiveTestUser(s.UserID)
		repo := &stubSessionRepo{getByTokenHashFn: func(context.Context, string) (*entity.Session, error) {
			return s, nil
		}}
		userRepo := &stubUserRepo{getByIDFn: func(context.Context, uuid.UUID) (*entity.User, error) {
			return u, nil
		}}
		auth := NewSessionAuthenticator(repo, userRepo, stubTokens{}, clk, cfg, log)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rr := httptest.NewRecorder()
		auth.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.EqualValues(t, 1, repo.updateCount.Load())
	})
}

func TestOptionalAuthenticate_PassesThroughOnFailure(t *testing.T) {
	clk := clock.MockClock{FixedTime: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	repo := &stubSessionRepo{getByTokenHashFn: func(context.Context, string) (*entity.Session, error) {
		return nil, nil
	}}
	auth := NewSessionAuthenticator(repo, &stubUserRepo{}, stubTokens{}, clk, testCfg(), discardLogger())

	called := false
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer unknown")
	rr := httptest.NewRecorder()
	auth.OptionalAuthenticate(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		_, ok := GetUserID(r.Context())
		assert.False(t, ok, "no user context expected when token is unknown")
	})).ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}
