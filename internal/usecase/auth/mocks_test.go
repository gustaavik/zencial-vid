package auth

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	infraAuth "github.com/zenfulcode/zencial/internal/infrastructure/auth"
)

// --- Mock UserRepository ---

type mockUserRepo struct {
	createFn        func(ctx context.Context, user *entity.User) error
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*entity.User, error)
	getByEmailFn    func(ctx context.Context, email valueobject.Email) (*entity.User, error)
	updateFn        func(ctx context.Context, user *entity.User) error
	deleteFn        func(ctx context.Context, id uuid.UUID) error
	existsByEmailFn func(ctx context.Context, email valueobject.Email) (bool, error)
	listFn          func(ctx context.Context, page, perPage int) ([]entity.User, int64, error)
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

func (m *mockUserRepo) List(ctx context.Context, page, perPage int) ([]entity.User, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, perPage)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.UserStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}

// --- Mock TokenService ---

type mockTokenService struct {
	generateTokenPairFn    func(userID uuid.UUID, role entity.UserRole) (*infraAuth.TokenPair, error)
	validateAccessTokenFn  func(tokenString string) (*infraAuth.Claims, error)
	generateRefreshTokenFn func() (string, error)
}

func (m *mockTokenService) GenerateTokenPair(userID uuid.UUID, role entity.UserRole) (*infraAuth.TokenPair, error) {
	if m.generateTokenPairFn != nil {
		return m.generateTokenPairFn(userID, role)
	}
	return &infraAuth.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}, nil
}

func (m *mockTokenService) ValidateAccessToken(tokenString string) (*infraAuth.Claims, error) {
	if m.validateAccessTokenFn != nil {
		return m.validateAccessTokenFn(tokenString)
	}
	return nil, nil
}

func (m *mockTokenService) GenerateRefreshToken() (string, error) {
	if m.generateRefreshTokenFn != nil {
		return m.generateRefreshTokenFn()
	}
	return "refresh-token", nil
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

// --- Mock SessionStore ---

type mockSessionStore struct {
	storeRefreshTokenFn       func(ctx context.Context, token string, userID uuid.UUID) error
	getUserIDByRefreshTokenFn func(ctx context.Context, token string) (uuid.UUID, error)
	deleteRefreshTokenFn      func(ctx context.Context, token string) error
	deleteAllUserTokensFn     func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockSessionStore) StoreRefreshToken(ctx context.Context, token string, userID uuid.UUID) error {
	if m.storeRefreshTokenFn != nil {
		return m.storeRefreshTokenFn(ctx, token, userID)
	}
	return nil
}

func (m *mockSessionStore) GetUserIDByRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	if m.getUserIDByRefreshTokenFn != nil {
		return m.getUserIDByRefreshTokenFn(ctx, token)
	}
	return uuid.Nil, nil
}

func (m *mockSessionStore) DeleteRefreshToken(ctx context.Context, token string) error {
	if m.deleteRefreshTokenFn != nil {
		return m.deleteRefreshTokenFn(ctx, token)
	}
	return nil
}

func (m *mockSessionStore) DeleteAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	if m.deleteAllUserTokensFn != nil {
		return m.deleteAllUserTokensFn(ctx, userID)
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

func newTestService(
	repo *mockUserRepo,
	tokenSvc *mockTokenService,
	hasher *mockHasher,
	sessionStore *mockSessionStore,
	dispatcher *mockDispatcher,
) *Service {
	if repo == nil {
		repo = &mockUserRepo{}
	}
	if tokenSvc == nil {
		tokenSvc = &mockTokenService{}
	}
	if hasher == nil {
		hasher = &mockHasher{}
	}
	if sessionStore == nil {
		sessionStore = &mockSessionStore{}
	}
	if dispatcher == nil {
		dispatcher = &mockDispatcher{}
	}
	return NewService(repo, tokenSvc, hasher, sessionStore, dispatcher, newTestLogger())
}
