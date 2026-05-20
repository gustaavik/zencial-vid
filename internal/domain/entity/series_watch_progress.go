package entity

import (
	"time"

	"github.com/google/uuid"
)

// SeriesWatchProgress tracks the last-watched episode for a user in a series.
type SeriesWatchProgress struct {
	UserID        uuid.UUID
	SeriesID      uuid.UUID
	LastEpisodeID uuid.UUID
	UpdatedAt     time.Time
}
