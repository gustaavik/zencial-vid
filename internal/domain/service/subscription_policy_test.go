package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func newTestPlan(name string, amount int64) *entity.Plan {
	return &entity.Plan{
		ID:              uuid.New(),
		Name:            name,
		Tier:            "basic",
		Price:           valueobject.NewMoney(amount, "USD"),
		BillingInterval: "monthly",
		MaxQuality:      valueobject.QualityHD,
		MaxStreams:      1,
		IsActive:        true,
		CreatedAt:       time.Now(),
	}
}

func TestSubscriptionPolicyService_CanChangePlan(t *testing.T) {
	svc := NewSubscriptionPolicyService()
	newPlan := newTestPlan("Standard", 1299)

	t.Run("nil subscription allows change", func(t *testing.T) {
		ok, reason := svc.CanChangePlan(nil, newPlan)
		assert.True(t, ok)
		assert.Empty(t, reason)
	})

	t.Run("expired subscription allows change", func(t *testing.T) {
		sub := &entity.Subscription{
			ID:     uuid.New(),
			PlanID: uuid.New(),
			Status: entity.SubscriptionExpired,
		}
		ok, reason := svc.CanChangePlan(sub, newPlan)
		assert.True(t, ok)
		assert.Empty(t, reason)
	})

	t.Run("same plan id is rejected", func(t *testing.T) {
		sub := &entity.Subscription{
			ID:                 uuid.New(),
			PlanID:             newPlan.ID,
			Status:             entity.SubscriptionActive,
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
			CurrentPeriodStart: time.Now(),
		}
		ok, reason := svc.CanChangePlan(sub, newPlan)
		assert.False(t, ok)
		assert.Contains(t, reason, "already subscribed")
	})

	t.Run("different plan is allowed", func(t *testing.T) {
		sub := &entity.Subscription{
			ID:                 uuid.New(),
			PlanID:             uuid.New(), // Different from newPlan.ID
			Status:             entity.SubscriptionActive,
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
			CurrentPeriodStart: time.Now(),
		}
		ok, reason := svc.CanChangePlan(sub, newPlan)
		assert.True(t, ok)
		assert.Empty(t, reason)
	})

	t.Run("inaccessible subscription allows change", func(t *testing.T) {
		sub := &entity.Subscription{
			ID:               uuid.New(),
			PlanID:           uuid.New(),
			Status:           entity.SubscriptionCanceled,
			CurrentPeriodEnd: time.Now().Add(-1 * time.Hour), // Past period end
		}
		ok, reason := svc.CanChangePlan(sub, newPlan)
		assert.True(t, ok)
		assert.Empty(t, reason)
	})
}

func TestSubscriptionPolicyService_DetermineChangeType(t *testing.T) {
	svc := NewSubscriptionPolicyService()

	t.Run("nil current plan returns created", func(t *testing.T) {
		result := svc.DetermineChangeType(nil, newTestPlan("Standard", 1299))
		assert.Equal(t, "created", result)
	})

	t.Run("higher price is upgrade", func(t *testing.T) {
		current := newTestPlan("Basic", 799)
		target := newTestPlan("Premium", 1799)
		result := svc.DetermineChangeType(current, target)
		assert.Equal(t, "upgraded", result)
	})

	t.Run("lower price is downgrade", func(t *testing.T) {
		current := newTestPlan("Premium", 1799)
		target := newTestPlan("Basic", 799)
		result := svc.DetermineChangeType(current, target)
		assert.Equal(t, "downgraded", result)
	})

	t.Run("same price is downgrade", func(t *testing.T) {
		current := newTestPlan("Standard", 1299)
		target := newTestPlan("Standard Alt", 1299)
		result := svc.DetermineChangeType(current, target)
		assert.Equal(t, "downgraded", result)
	})
}
