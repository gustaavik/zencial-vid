package dto

// PlanResponse represents a subscription plan in API responses.
type PlanResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name" example:"Premium"`
	Tier             string `json:"tier" example:"premium"`
	PriceAmount      int64  `json:"price_amount" example:"1499"`
	PriceCurrency    string `json:"price_currency" example:"USD"`
	BillingInterval  string `json:"billing_interval" example:"monthly"`
	MaxQuality       string `json:"max_quality" example:"UHD"`
	MaxStreams       int    `json:"max_streams" example:"4"`
	DownloadsAllowed bool   `json:"downloads_allowed"`
	IsActive         bool   `json:"is_active"`
}

// SubscriptionResponse represents a subscription in API responses.
type SubscriptionResponse struct {
	ID                 string       `json:"id"`
	Plan               PlanResponse `json:"plan"`
	Status             string       `json:"status" example:"active"`
	CurrentPeriodStart string       `json:"current_period_start"`
	CurrentPeriodEnd   string       `json:"current_period_end"`
	CanceledAt         *string      `json:"canceled_at,omitempty"`
	CreatedAt          string       `json:"created_at"`
}

// SubscribeRequest represents a subscription creation request.
type SubscribeRequest struct {
	PlanID        string `json:"plan_id" validate:"required,uuid"`
	PaymentMethod string `json:"payment_method" validate:"required"`
}

// ChangePlanRequest represents a plan change request.
type ChangePlanRequest struct {
	PlanID string `json:"plan_id" validate:"required,uuid"`
}

// AdminSubscriptionResponse represents a subscription with user info for admin views.
type AdminSubscriptionResponse struct {
	ID                 string       `json:"id"`
	UserID             string       `json:"user_id"`
	UserEmail          string       `json:"user_email"`
	Plan               PlanResponse `json:"plan"`
	Status             string       `json:"status" example:"active"`
	CurrentPeriodStart string       `json:"current_period_start"`
	CurrentPeriodEnd   string       `json:"current_period_end"`
	CanceledAt         *string      `json:"canceled_at,omitempty"`
	CreatedAt          string       `json:"created_at"`
}

// AdminChangePlanRequest represents an admin plan change request.
type AdminChangePlanRequest struct {
	PlanID string `json:"plan_id" validate:"required,uuid"`
}

// AdminCreateSubscriptionRequest represents an admin request to assign a subscription to a user.
type AdminCreateSubscriptionRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	PlanID string `json:"plan_id" validate:"required,uuid"`
}

// CreatePlanRequest represents a request to create a new subscription plan.
type CreatePlanRequest struct {
	Name             string `json:"name" validate:"required"`
	Tier             string `json:"tier" validate:"required"`
	PriceAmount      int64  `json:"price_amount" validate:"required,min=0"`
	PriceCurrency    string `json:"price_currency" validate:"required"`
	BillingInterval  string `json:"billing_interval" validate:"required,oneof=monthly yearly"`
	MaxQuality       string `json:"max_quality" validate:"required,oneof=SD HD FHD UHD"`
	MaxStreams       int    `json:"max_streams" validate:"required,min=1"`
	DownloadsAllowed bool   `json:"downloads_allowed"`
}

// UpdatePlanRequest represents a request to update an existing subscription plan.
type UpdatePlanRequest struct {
	Name             string `json:"name" validate:"required"`
	Tier             string `json:"tier" validate:"required"`
	PriceAmount      int64  `json:"price_amount" validate:"min=0"`
	PriceCurrency    string `json:"price_currency" validate:"required"`
	BillingInterval  string `json:"billing_interval" validate:"required,oneof=monthly yearly"`
	MaxQuality       string `json:"max_quality" validate:"required,oneof=SD HD FHD UHD"`
	MaxStreams       int    `json:"max_streams" validate:"required,min=1"`
	DownloadsAllowed bool   `json:"downloads_allowed"`
	IsActive         bool   `json:"is_active"`
}
