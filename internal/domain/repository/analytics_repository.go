package repository

import (
	"context"

	"github.com/google/uuid"
)

// VideoStats holds aggregated viewing statistics for a single video.
type VideoStats struct {
	VideoID        uuid.UUID
	TotalViewers   int64
	AvgProgressPct float64
	CompletionRate float64 // percentage of viewers who reached >= 90% of duration
}

// AnalyticsRepository defines read-only aggregate queries over watch progress data.
type AnalyticsRepository interface {
	GetVideoStats(ctx context.Context, videoID uuid.UUID) (*VideoStats, error)
	GetUploaderSummary(ctx context.Context, uploaderID uuid.UUID) ([]VideoStats, error)
}
