package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// SeriesWatchProgressRepository defines persistence for series-level watch progress.
type SeriesWatchProgressRepository interface {
	Upsert(ctx context.Context, userID, seriesID, lastEpisodeID uuid.UUID) error
	Get(ctx context.Context, userID, seriesID uuid.UUID) (*entity.SeriesWatchProgress, error)
	Delete(ctx context.Context, userID, seriesID uuid.UUID) error
}
