package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// WatchProgressRepository defines persistence operations for video playback progress.
type WatchProgressRepository interface {
	// Upsert creates or updates the user's progress for the given video.
	Upsert(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int64) error
	// Get returns the user's progress for the given video, or (nil, nil) if none exists.
	Get(ctx context.Context, userID, videoID uuid.UUID) (*entity.WatchProgress, error)
	// Delete removes the user's progress for the given video.
	Delete(ctx context.Context, userID, videoID uuid.UUID) error
	// ListInProgress returns published videos the user has started but not finished
	// (position < 95% of duration), ordered by most recently updated.
	ListInProgress(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.VideoWithProgress, int64, error)
}
