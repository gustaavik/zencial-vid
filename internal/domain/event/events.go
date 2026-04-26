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

// VideoTranscodeFailed is emitted when CDN transcoding fails for a video.
type VideoTranscodeFailed struct {
	VideoID   uuid.UUID
	Reason    string
	Timestamp time.Time
}

func (e VideoTranscodeFailed) EventName() string     { return "video.transcode_failed" }
func (e VideoTranscodeFailed) OccurredAt() time.Time { return e.Timestamp }

// VideoArchived is emitted when a video is soft-deleted (archived).
type VideoArchived struct {
	VideoID   uuid.UUID
	Timestamp time.Time
}

func (e VideoArchived) EventName() string     { return "video.archived" }
func (e VideoArchived) OccurredAt() time.Time { return e.Timestamp }

// VideoRestored is emitted when an archived video is restored.
type VideoRestored struct {
	VideoID   uuid.UUID
	Timestamp time.Time
}

func (e VideoRestored) EventName() string     { return "video.restored" }
func (e VideoRestored) OccurredAt() time.Time { return e.Timestamp }

// UserProfileUpdated is emitted when a user updates their profile.
type UserProfileUpdated struct {
	UserID    uuid.UUID
	Timestamp time.Time
}

func (e UserProfileUpdated) EventName() string     { return "user.profile_updated" }
func (e UserProfileUpdated) OccurredAt() time.Time { return e.Timestamp }

// UserAccountDeleted is emitted when a user soft-deletes their account.
type UserAccountDeleted struct {
	UserID    uuid.UUID
	Timestamp time.Time
}

func (e UserAccountDeleted) EventName() string     { return "user.account_deleted" }
func (e UserAccountDeleted) OccurredAt() time.Time { return e.Timestamp }

// UserStatusChanged is emitted when an admin changes a user's status.
type UserStatusChanged struct {
	UserID    uuid.UUID
	NewStatus string
	Timestamp time.Time
}

func (e UserStatusChanged) EventName() string     { return "user.status_changed" }
func (e UserStatusChanged) OccurredAt() time.Time { return e.Timestamp }
