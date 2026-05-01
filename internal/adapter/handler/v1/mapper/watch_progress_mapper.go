package mapper

import (
	"context"
	"math"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
)

const timeFormat = "2006-01-02T15:04:05Z"

// percentWatched returns position/duration as a percentage rounded to two decimals.
// Returns 0 for zero or negative durations.
func percentWatched(positionSeconds, durationSeconds int64) float64 {
	if durationSeconds <= 0 {
		return 0
	}
	pct := float64(positionSeconds) / float64(durationSeconds) * 100
	return math.Round(pct*100) / 100
}

// WatchProgressToResponse maps a WatchProgress entity (with the video's duration)
// to the API response DTO.
func WatchProgressToResponse(progress *entity.WatchProgress, durationSeconds int64) dto.WatchProgressResponse {
	return dto.WatchProgressResponse{
		VideoID:         progress.VideoID.String(),
		PositionSeconds: progress.PositionSeconds,
		DurationSeconds: durationSeconds,
		PercentWatched:  percentWatched(progress.PositionSeconds, durationSeconds),
		UpdatedAt:       progress.UpdatedAt.Format(timeFormat),
	}
}

// ContinueWatchingItemsToResponse maps a slice of VideoWithProgress entities
// to the "continue watching" feed DTO.
func ContinueWatchingItemsToResponse(ctx context.Context, items []entity.VideoWithProgress, store storage.StorageService) []dto.ContinueWatchingItem {
	result := make([]dto.ContinueWatchingItem, len(items))
	for i := range items {
		duration := items[i].Video.Duration.Seconds
		result[i] = dto.ContinueWatchingItem{
			Video:           VideoToResponse(ctx, &items[i].Video, store),
			PositionSeconds: items[i].Progress.PositionSeconds,
			DurationSeconds: duration,
			PercentWatched:  percentWatched(items[i].Progress.PositionSeconds, duration),
			UpdatedAt:       items[i].Progress.UpdatedAt.Format(timeFormat),
		}
	}
	return result
}
