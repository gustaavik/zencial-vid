package subscription

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles subscription use cases.
type Service struct {
	subscriptionRepo repository.SubscriptionRepository
	log              *slog.Logger
}

// NewService creates a new subscription Service.
func NewService(subscriptionRepo repository.SubscriptionRepository, log *slog.Logger) *Service {
	return &Service{subscriptionRepo: subscriptionRepo, log: log}
}
