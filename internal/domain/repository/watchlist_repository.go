package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// WatchlistRepository defines persistence operations for user watchlists.
type WatchlistRepository interface {
	Add(ctx context.Context, userID, videoID uuid.UUID) error
	Remove(ctx context.Context, userID, videoID uuid.UUID) error
	Exists(ctx context.Context, userID, videoID uuid.UUID) (bool, error)
	ListVideos(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.Video, int64, error)
}
