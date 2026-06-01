package entity

import (
	"time"

	"github.com/google/uuid"
)

// VideoCast represents a cast member's credit for a specific video
// (role and sort position within that video).
type VideoCast struct {
	ID        uuid.UUID
	VideoID   uuid.UUID
	CastID    uuid.UUID
	Role      string
	SortOrder int
	Cast      *Cast  // populated by repository JOIN for ListByVideo direction
	Video     *Video // populated by repository JOIN for ListByCast direction
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewVideoCast creates a new VideoCast record linking a cast member to a video.
func NewVideoCast(videoID, castID uuid.UUID, role string, sortOrder int) *VideoCast {
	now := time.Now().UTC()
	return &VideoCast{
		ID:        uuid.Must(uuid.NewV7()),
		VideoID:   videoID,
		CastID:    castID,
		Role:      role,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
