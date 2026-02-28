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
