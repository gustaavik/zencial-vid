package dto

// PlanResponse represents a plan in API responses.
type PlanResponse struct {
	ID            string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          string  `json:"name" example:"Premium"`
	Slug          string  `json:"slug" example:"premium"`
	Description   string  `json:"description" example:"Access to all premium content"`
	Price         float64 `json:"price" example:"9.99"`
	Level         int     `json:"level" example:"2"`
	StripePriceID string  `json:"stripe_price_id,omitempty" example:"price_123"`
	IsActive      bool    `json:"is_active" example:"true"`
	CreatedAt     string  `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt     string  `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// CreatePlanRequest represents a plan creation request.
type CreatePlanRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=255"`
	Description   string  `json:"description" validate:"max=2000"`
	Price         float64 `json:"price" validate:"gte=0"`
	Level         int     `json:"level" validate:"gte=0"`
	StripePriceID string  `json:"stripe_price_id,omitempty" validate:"omitempty,max=255"`
}

// UpdatePlanRequest represents a plan update request.
type UpdatePlanRequest struct {
	Name          *string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description   *string  `json:"description,omitempty" validate:"omitempty,max=2000"`
	Price         *float64 `json:"price,omitempty" validate:"omitempty,gte=0"`
	Level         *int     `json:"level,omitempty" validate:"omitempty,gte=0"`
	StripePriceID *string  `json:"stripe_price_id,omitempty" validate:"omitempty,max=255"`
	IsActive      *bool    `json:"is_active,omitempty"`
}
