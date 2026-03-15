package dto

// SubscriptionResponse represents a subscription in API responses.
type SubscriptionResponse struct {
	ID        string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string  `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PlanID    string  `json:"plan_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PlanName  string  `json:"plan_name,omitempty" example:"Premium"`
	PlanLevel int     `json:"plan_level,omitempty" example:"2"`
	Status    string  `json:"status" example:"active"`
	StartedAt string  `json:"started_at" example:"2025-01-01T00:00:00Z"`
	ExpiresAt *string `json:"expires_at,omitempty" example:"2025-12-31T23:59:59Z"`
	CreatedAt string  `json:"created_at" example:"2025-01-01T00:00:00Z"`
}

// AssignSubscriptionRequest represents a subscription assignment request.
type AssignSubscriptionRequest struct {
	UserID    string  `json:"user_id" validate:"required,uuid"`
	PlanID    string  `json:"plan_id" validate:"required,uuid"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}
