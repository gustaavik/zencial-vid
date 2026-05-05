package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog records a single domain event for admin observability.
// Captures who did what, on which entity, and when. Metadata holds
// event-specific context as JSON; before/after diffs are out of scope.
type AuditLog struct {
	ID         uuid.UUID
	ActorID    *uuid.UUID // nil = system-initiated
	EventName  string     // e.g. "user.role_changed"
	EntityType string     // "user" | "video" | "genre" | "plan" | "subscription"
	EntityID   *uuid.UUID
	Metadata   map[string]any
	OccurredAt time.Time
	CreatedAt  time.Time

	// ActorEmail is denormalized for list responses; populated by repository
	// queries that join users. Empty when ActorID is nil or actor was deleted.
	ActorEmail string
}
