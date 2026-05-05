package subscription

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles subscription use cases.
type Service struct {
	subRepo    repository.SubscriptionRepository
	planRepo   repository.PlanRepository
	dispatcher event.Dispatcher
	log        *slog.Logger
}

// NewService creates a new subscription Service.
func NewService(
	subRepo repository.SubscriptionRepository,
	planRepo repository.PlanRepository,
	dispatcher event.Dispatcher,
	log *slog.Logger,
) *Service {
	return &Service{subRepo: subRepo, planRepo: planRepo, dispatcher: dispatcher, log: log}
}
