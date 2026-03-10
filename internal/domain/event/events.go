package event

import (
	"time"

	"github.com/google/uuid"
)

// Event is the interface that all domain events implement.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// UserRegistered is emitted when a new user registers.
type UserRegistered struct {
	UserID    uuid.UUID
	Email     string
	Timestamp time.Time
}

func (e UserRegistered) EventName() string     { return "user.registered" }
func (e UserRegistered) OccurredAt() time.Time { return e.Timestamp }

// UserLoggedIn is emitted when a user logs in.
type UserLoggedIn struct {
	UserID    uuid.UUID
	Timestamp time.Time
}

func (e UserLoggedIn) EventName() string     { return "user.logged_in" }
func (e UserLoggedIn) OccurredAt() time.Time { return e.Timestamp }

// VideoUploaded is emitted when a new video is uploaded.
type VideoUploaded struct {
	VideoID    uuid.UUID
	Title      string
	UploadedBy uuid.UUID
	Timestamp  time.Time
}

func (e VideoUploaded) EventName() string     { return "video.uploaded" }
func (e VideoUploaded) OccurredAt() time.Time { return e.Timestamp }

// VideoPublished is emitted when a video is published.
type VideoPublished struct {
	VideoID   uuid.UUID
	Timestamp time.Time
}

func (e VideoPublished) EventName() string     { return "video.published" }
func (e VideoPublished) OccurredAt() time.Time { return e.Timestamp }
