package session

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// mockSessionRepo is a closure-based test double for SessionRepository.
type mockSessionRepo struct {
	createFn              func(ctx context.Context, s *entity.Session) error
	getByTokenHashFn      func(ctx context.Context, hash string) (*entity.Session, error)
	getByIDFn             func(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	listByUserIDFn        func(ctx context.Context, userID uuid.UUID, fs *filter.FilterSet) ([]entity.Session, int64, error)
	updateActivityFn      func(ctx context.Context, id uuid.UUID, last, idle time.Time) error
	revokeFn              func(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	revokeAllForUserFn    func(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error)
	revokeOthersForUserFn func(ctx context.Context, userID, exceptID uuid.UUID, revokedAt time.Time) (int64, error)
	deleteExpiredFn       func(ctx context.Context, before time.Time) (int64, error)
}

func (m *mockSessionRepo) Create(ctx context.Context, s *entity.Session) error {
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
func (m *mockSessionRepo) UpdateActivity(ctx context.Context, id uuid.UUID, last, idle time.Time) error {
	if m.updateActivityFn != nil {
		return m.updateActivityFn(ctx, id, last, idle)
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

type mockDispatcher struct {
	dispatched []event.Event
}

func (m *mockDispatcher) Dispatch(evt event.Event) error {
	m.dispatched = append(m.dispatched, evt)
	return nil
}
func (m *mockDispatcher) Subscribe(string, func(event.Event) error) {}
func (m *mockDispatcher) SubscribeAll(func(event.Event) error)      {}

var fixedNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func newTestService(repo *mockSessionRepo, dispatcher *mockDispatcher) *Service {
	if repo == nil {
		repo = &mockSessionRepo{}
	}
	if dispatcher == nil {
		dispatcher = &mockDispatcher{}
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewService(repo, dispatcher, log, clock.MockClock{FixedTime: fixedNow})
}
