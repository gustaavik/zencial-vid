package user

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles user profile use cases.
type Service struct {
	userRepo   repository.UserRepository
	dispatcher event.Dispatcher
	log        *slog.Logger
}

// NewService creates a new user Service.
func NewService(
	userRepo repository.UserRepository,
	dispatcher event.Dispatcher,
	log *slog.Logger,
) *Service {
	return &Service{
		userRepo:   userRepo,
		dispatcher: dispatcher,
		log:        log,
	}
}
