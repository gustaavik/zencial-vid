package entity

import (
	"time"

	"github.com/google/uuid"
)

// CastStatus represents the lifecycle state of a cast member.
type CastStatus string

const (
	CastStatusActive   CastStatus = "active"
	CastStatusArchived CastStatus = "archived"
)

// Cast is a standalone cast member (actor, director, writer, etc.) that can
// be credited on multiple videos.
type Cast struct {
	ID         uuid.UUID
	Name       string
	Status     CastStatus
	PictureKey string // stored in DB; empty if no picture assigned
	PictureURL string // resolved by use case from PictureKey; not stored in DB
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewCast creates a new Cast record.
func NewCast(name string) *Cast {
	now := time.Now().UTC()
	return &Cast{
		ID:        uuid.Must(uuid.NewV7()),
		Name:      name,
		Status:    CastStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Archive soft-deletes the cast member.
func (c *Cast) Archive() {
	c.Status = CastStatusArchived
	c.UpdatedAt = time.Now().UTC()
}

// Unarchive restores an archived cast member to active status.
func (c *Cast) Unarchive() {
	c.Status = CastStatusActive
	c.UpdatedAt = time.Now().UTC()
}

// IsArchived reports whether the cast member is archived.
func (c *Cast) IsArchived() bool {
	return c.Status == CastStatusArchived
}
