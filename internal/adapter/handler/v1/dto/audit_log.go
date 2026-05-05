package dto

// AuditActorRef identifies the user who triggered an audit event. Nil in
// AuditLogResponse when the event was system-initiated.
type AuditActorRef struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// AuditLogResponse is the API representation of a single audit log entry.
type AuditLogResponse struct {
	ID         string         `json:"id"`
	Actor      *AuditActorRef `json:"actor"`
	EventName  string         `json:"event_name"`
	EntityType string         `json:"entity_type"`
	EntityID   *string        `json:"entity_id"`
	Metadata   map[string]any `json:"metadata"`
	OccurredAt string         `json:"occurred_at"`
}
