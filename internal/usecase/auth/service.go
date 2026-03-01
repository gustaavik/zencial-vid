package auth

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
)

// Service handles authentication use cases.
type Service struct {
	userRepo     repository.UserRepository
	tokenService auth.TokenService
	hasher       auth.PasswordHasher
	sessionStore repository.SessionStore
	dispatcher   event.Dispatcher
	log          *slog.Logger
}

// NewService creates a new auth Service.
func NewService(
	userRepo repository.UserRepository,
	tokenService auth.TokenService,
	hasher auth.PasswordHasher,
	sessionStore repository.SessionStore,
	dispatcher event.Dispatcher,
	log *slog.Logger,
) *Service {
	return &Service{
		userRepo:     userRepo,
		tokenService: tokenService,
		hasher:       hasher,
		sessionStore: sessionStore,
		dispatcher:   dispatcher,
		log:          log,
	}
}
