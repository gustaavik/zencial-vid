package entity

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus represents the lifecycle state of a subscription.
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
)

// Subscription represents a user's subscription to a plan.
type Subscription struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	PlanID               uuid.UUID
	Status               SubscriptionStatus
	StripeSubscriptionID string
	StripeCustomerID     string
	StartedAt            time.Time
	ExpiresAt            *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewSubscription creates a new active Subscription.
func NewSubscription(userID, planID uuid.UUID, expiresAt *time.Time) *Subscription {
	now := time.Now().UTC()
	return &Subscription{
		ID:        uuid.New(),
		UserID:    userID,
		PlanID:    planID,
		Status:    SubscriptionStatusActive,
		StartedAt: now,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsActive reports whether the subscription is currently active.
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive
}

// Cancel marks the subscription as cancelled.
func (s *Subscription) Cancel() {
	s.Status = SubscriptionStatusCancelled
	s.UpdatedAt = time.Now().UTC()
}
