package service

import (
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// SubscriptionPolicyService enforces subscription change rules.
type SubscriptionPolicyService struct{}

// NewSubscriptionPolicyService creates a new SubscriptionPolicyService.
func NewSubscriptionPolicyService() *SubscriptionPolicyService {
	return &SubscriptionPolicyService{}
}

// CanChangePlan checks if a user can switch from their current plan to a new plan.
func (s *SubscriptionPolicyService) CanChangePlan(current *entity.Subscription, newPlan *entity.Plan) (bool, string) {
	if current == nil {
		return true, ""
	}

	if current.Status == entity.SubscriptionExpired {
		return true, ""
	}

	if !current.IsAccessible() {
		return true, ""
	}

	if current.PlanID == newPlan.ID {
		return false, "already subscribed to this plan"
	}

	return true, ""
}

// DetermineChangeType returns whether a plan change is an upgrade or downgrade.
func (s *SubscriptionPolicyService) DetermineChangeType(currentPlan, newPlan *entity.Plan) string {
	if currentPlan == nil {
		return "created"
	}
	if newPlan.Price.Amount > currentPlan.Price.Amount {
		return "upgraded"
	}
	return "downgraded"
}
