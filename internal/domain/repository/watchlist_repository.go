package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// WatchlistRepository defines persistence operations for watchlists.
type WatchlistRepository interface {
	Add(ctx context.Context, item *entity.WatchlistItem) error
	Remove(ctx context.Context, userID, contentID uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID, page, perPage int) ([]entity.WatchlistItem, int64, error)
	Exists(ctx context.Context, userID, contentID uuid.UUID) (bool, error)
}
