package analytics

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles analytics use cases.
type Service struct {
	analyticsRepo repository.AnalyticsRepository
	videoRepo     repository.VideoRepository
	log           *slog.Logger
}

// NewService creates a new analytics Service.
func NewService(analyticsRepo repository.AnalyticsRepository, videoRepo repository.VideoRepository, log *slog.Logger) *Service {
	return &Service{analyticsRepo: analyticsRepo, videoRepo: videoRepo, log: log}
}
