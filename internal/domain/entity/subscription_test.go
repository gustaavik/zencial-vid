package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func newTestSubscription(status SubscriptionStatus, periodEnd time.Time) *Subscription {
	return &Subscription{
		ID:                 uuid.New(),
		UserID:             uuid.New(),
		PlanID:             uuid.New(),
		Status:             status,
		CurrentPeriodStart: time.Now().Add(-30 * 24 * time.Hour),
		CurrentPeriodEnd:   periodEnd,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func TestSubscription_IsAccessible(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name   string
		sub    *Subscription
		expect bool
	}{
		{
			name:   "active subscription is accessible",
			sub:    newTestSubscription(SubscriptionActive, future),
			expect: true,
		},
		{
			name:   "trialing subscription is accessible",
			sub:    newTestSubscription(SubscriptionTrialing, future),
			expect: true,
		},
		{
			name:   "canceled subscription before period end is accessible",
			sub:    newTestSubscription(SubscriptionCanceled, future),
			expect: true,
		},
		{
			name:   "canceled subscription after period end is not accessible",
			sub:    newTestSubscription(SubscriptionCanceled, past),
			expect: false,
		},
		{
			name:   "expired subscription is not accessible",
			sub:    newTestSubscription(SubscriptionExpired, future),
			expect: false,
		},
		{
			name:   "past due subscription is not accessible",
			sub:    newTestSubscription(SubscriptionPastDue, future),
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.sub.IsAccessible())
		})
	}
}

func TestSubscription_MaxVideoQuality(t *testing.T) {
	t.Run("nil plan returns SD", func(t *testing.T) {
		sub := newTestSubscription(SubscriptionActive, time.Now().Add(time.Hour))
		sub.Plan = nil
		assert.Equal(t, valueobject.QualitySD, sub.MaxVideoQuality())
	})

	t.Run("plan with FHD returns FHD", func(t *testing.T) {
		sub := newTestSubscription(SubscriptionActive, time.Now().Add(time.Hour))
		sub.Plan = &Plan{
			ID:         uuid.New(),
			Name:       "Standard",
			MaxQuality: valueobject.QualityFHD,
		}
		assert.Equal(t, valueobject.QualityFHD, sub.MaxVideoQuality())
	})

	t.Run("plan with UHD returns UHD", func(t *testing.T) {
		sub := newTestSubscription(SubscriptionActive, time.Now().Add(time.Hour))
		sub.Plan = &Plan{
			ID:         uuid.New(),
			Name:       "Premium",
			MaxQuality: valueobject.QualityUHD,
		}
		assert.Equal(t, valueobject.QualityUHD, sub.MaxVideoQuality())
	})
}

func TestSubscription_Cancel(t *testing.T) {
	sub := newTestSubscription(SubscriptionActive, time.Now().Add(30*24*time.Hour))
	beforeUpdate := sub.UpdatedAt

	time.Sleep(time.Millisecond)
	sub.Cancel()

	assert.Equal(t, SubscriptionCanceled, sub.Status)
	require.NotNil(t, sub.CanceledAt)
	assert.True(t, sub.UpdatedAt.After(beforeUpdate), "UpdatedAt should be updated")
}
