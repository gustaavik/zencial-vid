package session

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
)

// Service exposes session-management use cases for both the user-facing UI
// (list/revoke own sessions) and the admin dashboard (manage another user's
// sessions). Background cleanup of expired rows is also a session use case.
type Service struct {
	sessionRepo repository.SessionRepository
	dispatcher  event.Dispatcher
	log         *slog.Logger
	clock       clock.Clock
}

// NewService constructs a Service.
func NewService(
	sessionRepo repository.SessionRepository,
	dispatcher event.Dispatcher,
	log *slog.Logger,
	clk clock.Clock,
) *Service {
	return &Service{
		sessionRepo: sessionRepo,
		dispatcher:  dispatcher,
		log:         log,
		clock:       clk,
	}
}
