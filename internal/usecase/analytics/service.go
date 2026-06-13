package analytics

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
)

// Service handles analytics use cases.
type Service struct {
	analyticsRepo repository.AnalyticsRepository
	playbackRepo  repository.PlaybackSessionRepository
	videoRepo     repository.VideoRepository
	log           *slog.Logger
	clock         clock.Clock
}

// NewService creates a new analytics Service.
func NewService(
	analyticsRepo repository.AnalyticsRepository,
	playbackRepo repository.PlaybackSessionRepository,
	videoRepo repository.VideoRepository,
	log *slog.Logger,
	clk clock.Clock,
) *Service {
	return &Service{
		analyticsRepo: analyticsRepo,
		playbackRepo:  playbackRepo,
		videoRepo:     videoRepo,
		log:           log,
		clock:         clk,
	}
}
