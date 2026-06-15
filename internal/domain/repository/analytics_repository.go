package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// PlaybackScope restricts aggregate queries to a single video, all videos by
// one uploader, or (with both fields nil) the whole platform.
type PlaybackScope struct {
	UploaderID *uuid.UUID
	VideoID    *uuid.UUID
}

// PlaybackTotals holds aggregate viewing statistics over a time window.
type PlaybackTotals struct {
	Views             int64
	WatchedSeconds    int64
	UniqueViewers     int64
	AvgPercentWatched float64
	FinishRate        float64 // percentage of views that reached >= 90% of duration
}

// DailyStat is one day of viewing activity.
type DailyStat struct {
	Day            time.Time
	Views          int64
	WatchedSeconds int64
}

// VideoRollup holds per-video aggregate statistics for dashboard tables.
type VideoRollup struct {
	VideoID           uuid.UUID
	Title             string
	Status            entity.VideoStatus
	Views             int64
	WatchedSeconds    int64
	AvgPercentWatched float64
	FinishRate        float64
}

// BreakdownItem is one slice of a categorical breakdown (source, country, platform).
type BreakdownItem struct {
	Key   string
	Views int64
}

// BreakdownDimension selects the categorical column for GetBreakdown.
type BreakdownDimension string

const (
	BreakdownSource   BreakdownDimension = "source"
	BreakdownCountry  BreakdownDimension = "country"
	BreakdownPlatform BreakdownDimension = "platform"
)

// AnalyticsRepository defines read-only aggregate queries over playback sessions.
// All windows are half-open: [from, to).
type AnalyticsRepository interface {
	GetTotals(ctx context.Context, scope PlaybackScope, from, to time.Time) (*PlaybackTotals, error)
	GetDailySeries(ctx context.Context, scope PlaybackScope, from, to time.Time) ([]DailyStat, error)
	GetTopVideos(ctx context.Context, uploaderID *uuid.UUID, from, to time.Time, limit int) ([]VideoRollup, error)
	// GetRetention returns entity.RetentionBuckets entries; entry i is the
	// percentage of view-qualified sessions that watched bucket i.
	GetRetention(ctx context.Context, videoID uuid.UUID, from, to time.Time) ([]float64, error)
	GetBreakdown(ctx context.Context, videoID uuid.UUID, dim BreakdownDimension, from, to time.Time) ([]BreakdownItem, error)
}
