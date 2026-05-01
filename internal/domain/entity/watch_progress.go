package entity

import (
	"time"

	"github.com/google/uuid"
)

// WatchProgress records how far a user has progressed through a video.
type WatchProgress struct {
	UserID          uuid.UUID
	VideoID         uuid.UUID
	PositionSeconds int64
	UpdatedAt       time.Time
}

// VideoWithProgress couples a Video with its WatchProgress for list endpoints,
// avoiding an N+1 round-trip when hydrating "continue watching" results.
type VideoWithProgress struct {
	Video    Video
	Progress WatchProgress
}
