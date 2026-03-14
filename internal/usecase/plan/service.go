package plan

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles plan use cases.
type Service struct {
	planRepo repository.PlanRepository
	log      *slog.Logger
}

// NewService creates a new plan Service.
func NewService(planRepo repository.PlanRepository, log *slog.Logger) *Service {
	return &Service{planRepo: planRepo, log: log}
}
