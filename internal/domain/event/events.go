package event

import (
	"time"

	"github.com/google/uuid"
)

// Entity type constants used by audit logging.
const (
	EntityUser         = "user"
	EntityVideo        = "video"
	EntityGenre        = "genre"
	EntityPlan         = "plan"
	EntitySubscription = "subscription"
)

// Event is the interface that all domain events implement.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// AuditableEvent is implemented by events that should be persisted in the
// audit log. Every domain event in this package implements it. Audit metadata
// is intentionally a small, action-specific snapshot — no before/after diffs.
type AuditableEvent interface {
	Event
	// AuditActor returns the user who triggered the event, or nil for
	// system-initiated events (e.g. CDN transcode callbacks).
	AuditActor() *uuid.UUID
	// AuditEntityType returns the kind of entity affected.
	AuditEntityType() string
	// AuditEntityID returns the affected entity's id, or nil if not applicable.
	AuditEntityID() *uuid.UUID
	// AuditMetadata returns event-specific fields to persist as JSON.
	AuditMetadata() map[string]any
}

// UserRegistered is emitted when a new user registers.
type UserRegistered struct {
	UserID    uuid.UUID
	Email     string
	ActorID   *uuid.UUID // nil = self-registration; set when an admin created the user
	Timestamp time.Time
}

func (e UserRegistered) EventName() string         { return "user.registered" }
func (e UserRegistered) OccurredAt() time.Time     { return e.Timestamp }
func (e UserRegistered) AuditActor() *uuid.UUID    { return e.ActorID }
func (e UserRegistered) AuditEntityType() string   { return EntityUser }
func (e UserRegistered) AuditEntityID() *uuid.UUID { return new(e.UserID) }
func (e UserRegistered) AuditMetadata() map[string]any {
	return map[string]any{"email": e.Email}
}

// UserLoggedIn is emitted when a user logs in.
type UserLoggedIn struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Timestamp time.Time
}

func (e UserLoggedIn) EventName() string         { return "user.logged_in" }
func (e UserLoggedIn) OccurredAt() time.Time     { return e.Timestamp }
func (e UserLoggedIn) AuditActor() *uuid.UUID    { return new(e.UserID) }
func (e UserLoggedIn) AuditEntityType() string   { return EntityUser }
func (e UserLoggedIn) AuditEntityID() *uuid.UUID { return new(e.UserID) }
func (e UserLoggedIn) AuditMetadata() map[string]any {
	return map[string]any{"session_id": e.SessionID.String()}
}

// UserLoggedOut is emitted when a user explicitly logs out (session revoked
// by the same user via /auth/logout).
type UserLoggedOut struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Timestamp time.Time
}

func (e UserLoggedOut) EventName() string         { return "user.logged_out" }
func (e UserLoggedOut) OccurredAt() time.Time     { return e.Timestamp }
func (e UserLoggedOut) AuditActor() *uuid.UUID    { return new(e.UserID) }
func (e UserLoggedOut) AuditEntityType() string   { return EntityUser }
func (e UserLoggedOut) AuditEntityID() *uuid.UUID { return new(e.UserID) }
func (e UserLoggedOut) AuditMetadata() map[string]any {
	return map[string]any{"session_id": e.SessionID.String()}
}

// SessionRevoked is emitted when a session is revoked (by the user managing
// their devices, or by an admin acting on a user's session).
type SessionRevoked struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	ActorID   *uuid.UUID // the actor that revoked it; nil = system
	Timestamp time.Time
}

func (e SessionRevoked) EventName() string         { return "session.revoked" }
func (e SessionRevoked) OccurredAt() time.Time     { return e.Timestamp }
func (e SessionRevoked) AuditActor() *uuid.UUID    { return e.ActorID }
func (e SessionRevoked) AuditEntityType() string   { return EntityUser }
func (e SessionRevoked) AuditEntityID() *uuid.UUID { return new(e.UserID) }
func (e SessionRevoked) AuditMetadata() map[string]any {
	return map[string]any{"session_id": e.SessionID.String()}
}

// VideoUploaded is emitted when a new video is uploaded.
type VideoUploaded struct {
	VideoID    uuid.UUID
	Title      string
	UploadedBy uuid.UUID
	Timestamp  time.Time
}

func (e VideoUploaded) EventName() string         { return "video.uploaded" }
func (e VideoUploaded) OccurredAt() time.Time     { return e.Timestamp }
func (e VideoUploaded) AuditActor() *uuid.UUID    { return new(e.UploadedBy) }
func (e VideoUploaded) AuditEntityType() string   { return EntityVideo }
func (e VideoUploaded) AuditEntityID() *uuid.UUID { return new(e.VideoID) }
func (e VideoUploaded) AuditMetadata() map[string]any {
	return map[string]any{"title": e.Title}
}

// VideoUpdated is emitted when an admin updates video metadata or thumbnail.
type VideoUpdated struct {
	VideoID   uuid.UUID
	ActorID   *uuid.UUID
	Field     string // "metadata" or "thumbnail"
	Timestamp time.Time
}

func (e VideoUpdated) EventName() string         { return "video.updated" }
func (e VideoUpdated) OccurredAt() time.Time     { return e.Timestamp }
func (e VideoUpdated) AuditActor() *uuid.UUID    { return e.ActorID }
func (e VideoUpdated) AuditEntityType() string   { return EntityVideo }
func (e VideoUpdated) AuditEntityID() *uuid.UUID { return new(e.VideoID) }
func (e VideoUpdated) AuditMetadata() map[string]any {
	return map[string]any{"field": e.Field}
}

// VideoPublished is emitted when a video is published (after CDN transcode).
type VideoPublished struct {
	VideoID   uuid.UUID
	Timestamp time.Time
}

func (e VideoPublished) EventName() string             { return "video.published" }
func (e VideoPublished) OccurredAt() time.Time         { return e.Timestamp }
func (e VideoPublished) AuditActor() *uuid.UUID        { return nil }
func (e VideoPublished) AuditEntityType() string       { return EntityVideo }
func (e VideoPublished) AuditEntityID() *uuid.UUID     { return new(e.VideoID) }
func (e VideoPublished) AuditMetadata() map[string]any { return map[string]any{} }

// VideoTranscodeFailed is emitted when CDN transcoding fails for a video.
type VideoTranscodeFailed struct {
	VideoID   uuid.UUID
	Reason    string
	Timestamp time.Time
}

func (e VideoTranscodeFailed) EventName() string         { return "video.transcode_failed" }
func (e VideoTranscodeFailed) OccurredAt() time.Time     { return e.Timestamp }
func (e VideoTranscodeFailed) AuditActor() *uuid.UUID    { return nil }
func (e VideoTranscodeFailed) AuditEntityType() string   { return EntityVideo }
func (e VideoTranscodeFailed) AuditEntityID() *uuid.UUID { return new(e.VideoID) }
func (e VideoTranscodeFailed) AuditMetadata() map[string]any {
	return map[string]any{"reason": e.Reason}
}

// VideoArchived is emitted when a video is soft-deleted (archived).
type VideoArchived struct {
	VideoID   uuid.UUID
	ActorID   *uuid.UUID
	Timestamp time.Time
}

func (e VideoArchived) EventName() string             { return "video.archived" }
func (e VideoArchived) OccurredAt() time.Time         { return e.Timestamp }
func (e VideoArchived) AuditActor() *uuid.UUID        { return e.ActorID }
func (e VideoArchived) AuditEntityType() string       { return EntityVideo }
func (e VideoArchived) AuditEntityID() *uuid.UUID     { return new(e.VideoID) }
func (e VideoArchived) AuditMetadata() map[string]any { return map[string]any{} }

// VideoRestored is emitted when an archived video is restored.
type VideoRestored struct {
	VideoID   uuid.UUID
	ActorID   *uuid.UUID
	Timestamp time.Time
}

func (e VideoRestored) EventName() string             { return "video.restored" }
func (e VideoRestored) OccurredAt() time.Time         { return e.Timestamp }
func (e VideoRestored) AuditActor() *uuid.UUID        { return e.ActorID }
func (e VideoRestored) AuditEntityType() string       { return EntityVideo }
func (e VideoRestored) AuditEntityID() *uuid.UUID     { return new(e.VideoID) }
func (e VideoRestored) AuditMetadata() map[string]any { return map[string]any{} }

// UserProfileUpdated is emitted when a user updates their profile.
type UserProfileUpdated struct {
	UserID    uuid.UUID
	ActorID   *uuid.UUID // nil-or-self for self-update; admin id when changed by admin
	Timestamp time.Time
}

func (e UserProfileUpdated) EventName() string             { return "user.profile_updated" }
func (e UserProfileUpdated) OccurredAt() time.Time         { return e.Timestamp }
func (e UserProfileUpdated) AuditActor() *uuid.UUID        { return e.ActorID }
func (e UserProfileUpdated) AuditEntityType() string       { return EntityUser }
func (e UserProfileUpdated) AuditEntityID() *uuid.UUID     { return new(e.UserID) }
func (e UserProfileUpdated) AuditMetadata() map[string]any { return map[string]any{} }

// UserAccountDeleted is emitted when a user account is soft-deleted.
type UserAccountDeleted struct {
	UserID    uuid.UUID
	ActorID   *uuid.UUID
	Timestamp time.Time
}

func (e UserAccountDeleted) EventName() string             { return "user.account_deleted" }
func (e UserAccountDeleted) OccurredAt() time.Time         { return e.Timestamp }
func (e UserAccountDeleted) AuditActor() *uuid.UUID        { return e.ActorID }
func (e UserAccountDeleted) AuditEntityType() string       { return EntityUser }
func (e UserAccountDeleted) AuditEntityID() *uuid.UUID     { return new(e.UserID) }
func (e UserAccountDeleted) AuditMetadata() map[string]any { return map[string]any{} }

// UserStatusChanged is emitted when an admin changes a user's status.
type UserStatusChanged struct {
	UserID    uuid.UUID
	ActorID   *uuid.UUID
	NewStatus string
	Timestamp time.Time
}

func (e UserStatusChanged) EventName() string         { return "user.status_changed" }
func (e UserStatusChanged) OccurredAt() time.Time     { return e.Timestamp }
func (e UserStatusChanged) AuditActor() *uuid.UUID    { return e.ActorID }
func (e UserStatusChanged) AuditEntityType() string   { return EntityUser }
func (e UserStatusChanged) AuditEntityID() *uuid.UUID { return new(e.UserID) }
func (e UserStatusChanged) AuditMetadata() map[string]any {
	return map[string]any{"new_status": e.NewStatus}
}

// UserRoleChanged is emitted when an admin changes a user's roles.
type UserRoleChanged struct {
	UserID    uuid.UUID
	ActorID   *uuid.UUID
	OldRoles  []string
	NewRoles  []string
	Timestamp time.Time
}

func (e UserRoleChanged) EventName() string         { return "user.role_changed" }
func (e UserRoleChanged) OccurredAt() time.Time     { return e.Timestamp }
func (e UserRoleChanged) AuditActor() *uuid.UUID    { return e.ActorID }
func (e UserRoleChanged) AuditEntityType() string   { return EntityUser }
func (e UserRoleChanged) AuditEntityID() *uuid.UUID { return new(e.UserID) }
func (e UserRoleChanged) AuditMetadata() map[string]any {
	return map[string]any{"old_roles": e.OldRoles, "new_roles": e.NewRoles}
}

// GenreCreated is emitted when an admin creates a genre.
type GenreCreated struct {
	GenreID   uuid.UUID
	ActorID   *uuid.UUID
	Name      string
	Timestamp time.Time
}

func (e GenreCreated) EventName() string         { return "genre.created" }
func (e GenreCreated) OccurredAt() time.Time     { return e.Timestamp }
func (e GenreCreated) AuditActor() *uuid.UUID    { return e.ActorID }
func (e GenreCreated) AuditEntityType() string   { return EntityGenre }
func (e GenreCreated) AuditEntityID() *uuid.UUID { return new(e.GenreID) }
func (e GenreCreated) AuditMetadata() map[string]any {
	return map[string]any{"name": e.Name}
}

// GenreUpdated is emitted when an admin updates a genre.
type GenreUpdated struct {
	GenreID   uuid.UUID
	ActorID   *uuid.UUID
	Name      string
	Timestamp time.Time
}

func (e GenreUpdated) EventName() string         { return "genre.updated" }
func (e GenreUpdated) OccurredAt() time.Time     { return e.Timestamp }
func (e GenreUpdated) AuditActor() *uuid.UUID    { return e.ActorID }
func (e GenreUpdated) AuditEntityType() string   { return EntityGenre }
func (e GenreUpdated) AuditEntityID() *uuid.UUID { return new(e.GenreID) }
func (e GenreUpdated) AuditMetadata() map[string]any {
	return map[string]any{"name": e.Name}
}

// GenreDeleted is emitted when an admin deletes a genre.
type GenreDeleted struct {
	GenreID   uuid.UUID
	ActorID   *uuid.UUID
	Timestamp time.Time
}

func (e GenreDeleted) EventName() string             { return "genre.deleted" }
func (e GenreDeleted) OccurredAt() time.Time         { return e.Timestamp }
func (e GenreDeleted) AuditActor() *uuid.UUID        { return e.ActorID }
func (e GenreDeleted) AuditEntityType() string       { return EntityGenre }
func (e GenreDeleted) AuditEntityID() *uuid.UUID     { return new(e.GenreID) }
func (e GenreDeleted) AuditMetadata() map[string]any { return map[string]any{} }

// PlanCreated is emitted when an admin creates a plan.
type PlanCreated struct {
	PlanID    uuid.UUID
	ActorID   *uuid.UUID
	Name      string
	Slug      string
	Timestamp time.Time
}

func (e PlanCreated) EventName() string         { return "plan.created" }
func (e PlanCreated) OccurredAt() time.Time     { return e.Timestamp }
func (e PlanCreated) AuditActor() *uuid.UUID    { return e.ActorID }
func (e PlanCreated) AuditEntityType() string   { return EntityPlan }
func (e PlanCreated) AuditEntityID() *uuid.UUID { return new(e.PlanID) }
func (e PlanCreated) AuditMetadata() map[string]any {
	return map[string]any{"name": e.Name, "slug": e.Slug}
}

// PlanUpdated is emitted when an admin updates a plan.
type PlanUpdated struct {
	PlanID    uuid.UUID
	ActorID   *uuid.UUID
	Name      string
	Slug      string
	Timestamp time.Time
}

func (e PlanUpdated) EventName() string         { return "plan.updated" }
func (e PlanUpdated) OccurredAt() time.Time     { return e.Timestamp }
func (e PlanUpdated) AuditActor() *uuid.UUID    { return e.ActorID }
func (e PlanUpdated) AuditEntityType() string   { return EntityPlan }
func (e PlanUpdated) AuditEntityID() *uuid.UUID { return new(e.PlanID) }
func (e PlanUpdated) AuditMetadata() map[string]any {
	return map[string]any{"name": e.Name, "slug": e.Slug}
}

// PlanDeleted is emitted when an admin deletes a plan.
type PlanDeleted struct {
	PlanID    uuid.UUID
	ActorID   *uuid.UUID
	Timestamp time.Time
}

func (e PlanDeleted) EventName() string             { return "plan.deleted" }
func (e PlanDeleted) OccurredAt() time.Time         { return e.Timestamp }
func (e PlanDeleted) AuditActor() *uuid.UUID        { return e.ActorID }
func (e PlanDeleted) AuditEntityType() string       { return EntityPlan }
func (e PlanDeleted) AuditEntityID() *uuid.UUID     { return new(e.PlanID) }
func (e PlanDeleted) AuditMetadata() map[string]any { return map[string]any{} }

// SubscriptionAssigned is emitted when an admin assigns a plan to a user.
type SubscriptionAssigned struct {
	SubscriptionID uuid.UUID
	UserID         uuid.UUID
	PlanID         uuid.UUID
	ActorID        *uuid.UUID
	Timestamp      time.Time
}

func (e SubscriptionAssigned) EventName() string         { return "subscription.assigned" }
func (e SubscriptionAssigned) OccurredAt() time.Time     { return e.Timestamp }
func (e SubscriptionAssigned) AuditActor() *uuid.UUID    { return e.ActorID }
func (e SubscriptionAssigned) AuditEntityType() string   { return EntitySubscription }
func (e SubscriptionAssigned) AuditEntityID() *uuid.UUID { return new(e.SubscriptionID) }
func (e SubscriptionAssigned) AuditMetadata() map[string]any {
	return map[string]any{"user_id": e.UserID.String(), "plan_id": e.PlanID.String()}
}

// SubscriptionCancelled is emitted when an admin cancels a subscription.
type SubscriptionCancelled struct {
	SubscriptionID uuid.UUID
	UserID         uuid.UUID
	ActorID        *uuid.UUID
	Timestamp      time.Time
}

func (e SubscriptionCancelled) EventName() string         { return "subscription.cancelled" }
func (e SubscriptionCancelled) OccurredAt() time.Time     { return e.Timestamp }
func (e SubscriptionCancelled) AuditActor() *uuid.UUID    { return e.ActorID }
func (e SubscriptionCancelled) AuditEntityType() string   { return EntitySubscription }
func (e SubscriptionCancelled) AuditEntityID() *uuid.UUID { return new(e.SubscriptionID) }
func (e SubscriptionCancelled) AuditMetadata() map[string]any {
	return map[string]any{"user_id": e.UserID.String()}
}
