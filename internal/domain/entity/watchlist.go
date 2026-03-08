package entity

import (
	"time"

	"github.com/google/uuid"
)

// WatchlistItem represents a content item in a user's watchlist.
type WatchlistItem struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ContentID uuid.UUID
	Content   *ContentSummary // Loaded when needed
	AddedAt   time.Time
}

// NewWatchlistItem creates a new watchlist item.
func NewWatchlistItem(userID, contentID uuid.UUID) *WatchlistItem {
	return &WatchlistItem{
		ID:        uuid.New(),
		UserID:    userID,
		ContentID: contentID,
		AddedAt:   time.Now(),
	}
}
