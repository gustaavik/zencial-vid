package entity

import (
	"time"

	"github.com/google/uuid"
)

// Cast is a standalone cast member (actor, director, writer, etc.) that can
// be credited on multiple videos.
type Cast struct {
	ID         uuid.UUID
	Name       string
	PictureKey string // stored in DB; empty if no picture assigned
	PictureURL string // resolved by use case from PictureKey; not stored in DB
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewCast creates a new Cast record.
func NewCast(name string) *Cast {
	now := time.Now().UTC()
	return &Cast{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
