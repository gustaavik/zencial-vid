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

// PlaybackStarted is emitted when a user starts streaming content.
type PlaybackStarted struct {
	UserID    uuid.UUID
	ContentID uuid.UUID
	EpisodeID *uuid.UUID
	Timestamp time.Time
}

func (e PlaybackStarted) EventName() string     { return "playback.started" }
func (e PlaybackStarted) OccurredAt() time.Time { return e.Timestamp }

// SubscriptionChanged is emitted when a subscription is created, changed, or canceled.
type SubscriptionChanged struct {
	UserID     uuid.UUID
	OldPlanID  *uuid.UUID
	NewPlanID  uuid.UUID
	ChangeType string // "created", "upgraded", "downgraded", "canceled"
	Timestamp  time.Time
}

func (e SubscriptionChanged) EventName() string     { return "subscription.changed" }
func (e SubscriptionChanged) OccurredAt() time.Time { return e.Timestamp }

// ContentPublished is emitted when content is published.
type ContentPublished struct {
	ContentID uuid.UUID
	Title     string
	Timestamp time.Time
}

func (e ContentPublished) EventName() string     { return "content.published" }
func (e ContentPublished) OccurredAt() time.Time { return e.Timestamp }
