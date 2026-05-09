package auth

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
)

// SessionContext is the device/network metadata captured at session creation.
// Handlers should populate this from the inbound HTTP request so the resulting
// session row is meaningful in the user's "Active sessions" UI.
type SessionContext struct {
	DeviceName string
	UserAgent  string
	IPAddress  string
}

// Service handles authentication use cases (login, register, logout).
type Service struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	tokens      auth.SessionTokenService
	hasher      auth.PasswordHasher
	dispatcher  event.Dispatcher
	log         *slog.Logger
	clock       clock.Clock
	cfg         config.SessionConfig
}

// NewService creates a new auth Service.
func NewService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	tokens auth.SessionTokenService,
	hasher auth.PasswordHasher,
	dispatcher event.Dispatcher,
	log *slog.Logger,
	clk clock.Clock,
	cfg config.SessionConfig,
) *Service {
	return &Service{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokens:      tokens,
		hasher:      hasher,
		dispatcher:  dispatcher,
		log:         log,
		clock:       clk,
		cfg:         cfg,
	}
}
