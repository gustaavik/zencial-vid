package user

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
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

// --- Mock Dispatcher ---

type mockDispatcher struct {
	dispatched []event.Event
}

func (m *mockDispatcher) Dispatch(evt event.Event) error {
	m.dispatched = append(m.dispatched, evt)
	return nil
}

func (m *mockDispatcher) Subscribe(_ string, _ func(event.Event) error) {}

// --- Test Helpers ---

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestService(repo *mockUserRepo, dispatcher *mockDispatcher) *Service {
	if repo == nil {
		repo = &mockUserRepo{}
	}
	if dispatcher == nil {
		dispatcher = &mockDispatcher{}
	}
	return NewService(repo, dispatcher, newTestLogger())
}

func newActiveUser() *entity.User {
	now := time.Now().UTC()
	id := uuid.New()
	return &entity.User{
		ID:           id,
		Email:        valueobject.EmailFromTrusted("user@example.com"),
		PasswordHash: valueobject.NewHashedPassword("hashed"),
		Role:         entity.RoleUser,
		Status:       entity.UserStatusActive,
		Profile: entity.UserProfile{
			UserID:      id,
			DisplayName: "Test User",
			Language:    "en",
			Country:     "US",
			UpdatedAt:   now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
