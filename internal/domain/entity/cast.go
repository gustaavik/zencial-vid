package entity

import (
	"time"

	"github.com/google/uuid"
)

// Cast represents a person credited on a video (actor, director, writer, etc.).
type Cast struct {
	ID         uuid.UUID
	VideoID    uuid.UUID
	Name       string
	Role       string
	SortOrder  int
	PictureKey string // stored in DB; empty if no picture assigned
	PictureURL string // resolved by use case from PictureKey; not stored in DB
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewCast creates a new Cast record.
func NewCast(videoID uuid.UUID, name, role string, sortOrder int) *Cast {
	now := time.Now().UTC()
	return &Cast{
		ID:        uuid.New(),
		VideoID:   videoID,
		Name:      name,
		Role:      role,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
