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
