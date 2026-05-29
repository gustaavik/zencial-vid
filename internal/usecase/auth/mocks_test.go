package auth

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// --- Mock UserRepository ---

type mockUserRepo struct {
	createFn        func(ctx context.Context, user *entity.User) error
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*entity.User, error)
	getByEmailFn    func(ctx context.Context, email valueobject.Email) (*entity.User, error)
	updateFn        func(ctx context.Context, user *entity.User) error
	deleteFn        func(ctx context.Context, id uuid.UUID) error
	existsByEmailFn func(ctx context.Context, email valueobject.Email) (bool, error)
	listFn          func(ctx context.Context, fs *filter.FilterSet) ([]entity.User, int64, error)
	updateStatusFn  func(ctx context.Context, id uuid.UUID, status entity.UserStatus) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email valueobject.Email) (*entity.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *entity.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error) {
	if m.existsByEmailFn != nil {
		return m.existsByEmailFn(ctx, email)
	}
	return false, nil
}

func (m *mockUserRepo) HandleExists(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockUserRepo) List(ctx context.Context, fs *filter.FilterSet) ([]entity.User, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, fs)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.UserStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}

// --- Mock SessionTokenService ---

type mockTokenGen struct {
	tokens   []string // pop next on each Generate
	hashes   []string
	idx      atomic.Int32
	failNext bool
	failErr  error
}

func (m *mockTokenGen) Generate() (string, string, error) {
	if m.failNext {
		return "", "", m.failErr
	}
	i := int(m.idx.Add(1)) - 1
	if i >= len(m.tokens) {
		// Default deterministic value so tests don't have to set up sequences.
		return "test-token", "test-hash", nil
	}
	return m.tokens[i], m.hashes[i], nil
}

func (m *mockTokenGen) Hash(token string) string {
	return token + ":hashed"
}

// --- Mock SessionRepository ---

type mockSessionRepo struct {
	createFn              func(ctx context.Context, s *entity.Session) error
	getByTokenHashFn      func(ctx context.Context, hash string) (*entity.Session, error)
	getByIDFn             func(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	listByUserIDFn        func(ctx context.Context, userID uuid.UUID, fs *filter.FilterSet) ([]entity.Session, int64, error)
	updateActivityFn      func(ctx context.Context, id uuid.UUID, lastActivity, idleExpires time.Time) error
	revokeFn              func(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	revokeAllForUserFn    func(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error)
	revokeOthersForUserFn func(ctx context.Context, userID, exceptID uuid.UUID, revokedAt time.Time) (int64, error)
	deleteExpiredFn       func(ctx context.Context, before time.Time) (int64, error)
	created               []*entity.Session
}

func (m *mockSessionRepo) Create(ctx context.Context, s *entity.Session) error {
	m.created = append(m.created, s)
	if m.createFn != nil {
		return m.createFn(ctx, s)
	}
	return nil
}

func (m *mockSessionRepo) GetByTokenHash(ctx context.Context, hash string) (*entity.Session, error) {
	if m.getByTokenHashFn != nil {
		return m.getByTokenHashFn(ctx, hash)
	}
	return nil, nil
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSessionRepo) ListByUserID(ctx context.Context, userID uuid.UUID, fs *filter.FilterSet) ([]entity.Session, int64, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID, fs)
	}
	return nil, 0, nil
}

func (m *mockSessionRepo) UpdateActivity(ctx context.Context, id uuid.UUID, lastActivity, idleExpires time.Time) error {
	if m.updateActivityFn != nil {
		return m.updateActivityFn(ctx, id, lastActivity, idleExpires)
	}
	return nil
}

func (m *mockSessionRepo) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	if m.revokeFn != nil {
		return m.revokeFn(ctx, id, revokedAt)
	}
	return nil
}

func (m *mockSessionRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	if m.revokeAllForUserFn != nil {
		return m.revokeAllForUserFn(ctx, userID, revokedAt)
	}
	return 0, nil
}

func (m *mockSessionRepo) RevokeOthersForUser(ctx context.Context, userID, exceptID uuid.UUID, revokedAt time.Time) (int64, error) {
	if m.revokeOthersForUserFn != nil {
		return m.revokeOthersForUserFn(ctx, userID, exceptID, revokedAt)
	}
	return 0, nil
}

func (m *mockSessionRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	if m.deleteExpiredFn != nil {
		return m.deleteExpiredFn(ctx, before)
	}
	return 0, nil
}

// --- Mock PasswordHasher ---

type mockHasher struct {
	hashFn    func(password string) (string, error)
	compareFn func(hashedPassword, password string) error
}

func (m *mockHasher) Hash(password string) (string, error) {
	if m.hashFn != nil {
		return m.hashFn(password)
	}
	return "hashed-password", nil
}

func (m *mockHasher) Compare(hashedPassword, password string) error {
	if m.compareFn != nil {
		return m.compareFn(hashedPassword, password)
	}
	return nil
}

// --- Mock Dispatcher ---

type mockDispatcher struct {
	dispatched []event.Event
}

func (m *mockDispatcher) Dispatch(evt event.Event) error {
	m.dispatched = append(m.dispatched, evt)
	return nil
}

func (m *mockDispatcher) Subscribe(_ string, _ func(event.Event) error) {}

func (m *mockDispatcher) SubscribeAll(_ func(event.Event) error) {}

// --- Test Helpers ---

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// fixedNow is the canonical clock value used by tests so created sessions
// have predictable timestamps.
var fixedNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func testSessionConfig() config.SessionConfig {
	return config.SessionConfig{
		IdleTimeout:     30 * 24 * time.Hour,
		AbsoluteTimeout: 90 * 24 * time.Hour,
		SlideDebounce:   5 * time.Minute,
		CleanupInterval: time.Hour,
	}
}

func newTestService(
	repo *mockUserRepo,
	sessionRepo *mockSessionRepo,
	tokens *mockTokenGen,
	hasher *mockHasher,
	dispatcher *mockDispatcher,
) *Service {
	if repo == nil {
		repo = &mockUserRepo{}
	}
	if sessionRepo == nil {
		sessionRepo = &mockSessionRepo{}
	}
	if tokens == nil {
		tokens = &mockTokenGen{}
	}
	if hasher == nil {
		hasher = &mockHasher{}
	}
	if dispatcher == nil {
		dispatcher = &mockDispatcher{}
	}
	return NewService(
		repo,
		sessionRepo,
		tokens,
		hasher,
		dispatcher,
		newTestLogger(),
		clock.MockClock{FixedTime: fixedNow},
		testSessionConfig(),
	)
}
