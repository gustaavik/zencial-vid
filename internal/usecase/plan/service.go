package plan

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles plan use cases.
type Service struct {
	planRepo   repository.PlanRepository
	dispatcher event.Dispatcher
	log        *slog.Logger
}

// NewService creates a new plan Service.
func NewService(planRepo repository.PlanRepository, dispatcher event.Dispatcher, log *slog.Logger) *Service {
	return &Service{planRepo: planRepo, dispatcher: dispatcher, log: log}
}
