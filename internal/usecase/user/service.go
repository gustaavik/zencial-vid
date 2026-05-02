package user

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
)

// Service handles user profile use cases.
type Service struct {
	userRepo   repository.UserRepository
	hasher     auth.PasswordHasher
	dispatcher event.Dispatcher
	log        *slog.Logger
}

// NewService creates a new user Service.
func NewService(
	userRepo repository.UserRepository,
	hasher auth.PasswordHasher,
	dispatcher event.Dispatcher,
	log *slog.Logger,
) *Service {
	return &Service{
		userRepo:   userRepo,
		hasher:     hasher,
		dispatcher: dispatcher,
		log:        log,
	}
}
