package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// PlanTier represents a subscription plan level.
type PlanTier string

const (
	PlanBasic    PlanTier = "basic"
	PlanStandard PlanTier = "standard"
	PlanPremium  PlanTier = "premium"
)

// Plan represents a subscription plan.
type Plan struct {
	ID               uuid.UUID
	Name             string
	Tier             PlanTier
	Price            valueobject.Money
	BillingInterval  string // "monthly", "yearly"
	MaxQuality       valueobject.VideoQuality
	MaxStreams        int
	DownloadsAllowed bool
	IsActive         bool
	CreatedAt        time.Time
}

// SubscriptionStatus represents the state of a subscription.
type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "active"
	SubscriptionCanceled SubscriptionStatus = "canceled"
	SubscriptionPastDue  SubscriptionStatus = "past_due"
	SubscriptionExpired  SubscriptionStatus = "expired"
	SubscriptionTrialing SubscriptionStatus = "trialing"
)

// Subscription represents a user's subscription to a plan.
type Subscription struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	PlanID             uuid.UUID
	Plan               *Plan
	Status             SubscriptionStatus
	ExternalID         string // Payment provider subscription ID
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CanceledAt         *time.Time
	TrialEnd           *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// IsAccessible returns whether the subscription grants content access.
func (s *Subscription) IsAccessible() bool {
	switch s.Status {
	case SubscriptionActive, SubscriptionTrialing:
		return true
	case SubscriptionCanceled:
		return time.Now().Before(s.CurrentPeriodEnd)
	default:
		return false
	}
}

// MaxVideoQuality returns the highest quality this subscription permits.
func (s *Subscription) MaxVideoQuality() valueobject.VideoQuality {
	if s.Plan == nil {
		return valueobject.QualitySD
	}
	return s.Plan.MaxQuality
}

// Cancel marks the subscription as canceled.
func (s *Subscription) Cancel() {
	now := time.Now()
	s.Status = SubscriptionCanceled
	s.CanceledAt = &now
	s.UpdatedAt = now
}
