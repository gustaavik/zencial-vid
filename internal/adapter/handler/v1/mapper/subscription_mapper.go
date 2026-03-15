package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	subscriptionuc "github.com/zenfulcode/zencial/internal/usecase/subscription"
)

// SubscriptionToResponse maps a Subscription entity to a SubscriptionResponse DTO.
func SubscriptionToResponse(sub *entity.Subscription) dto.SubscriptionResponse {
	resp := dto.SubscriptionResponse{
		ID:        sub.ID.String(),
		UserID:    sub.UserID.String(),
		PlanID:    sub.PlanID.String(),
		Status:    string(sub.Status),
		StartedAt: sub.StartedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt: sub.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if sub.ExpiresAt != nil {
		expires := sub.ExpiresAt.Format("2006-01-02T15:04:05Z")
		resp.ExpiresAt = &expires
	}
	return resp
}

// SubscriptionWithPlanToResponse maps a SubscriptionWithPlan to a SubscriptionResponse DTO.
func SubscriptionWithPlanToResponse(swp *subscriptionuc.SubscriptionWithPlan) dto.SubscriptionResponse {
	resp := SubscriptionToResponse(swp.Subscription)
	if swp.Plan != nil {
		resp.PlanName = swp.Plan.Name
		resp.PlanLevel = swp.Plan.Level
	}
	return resp
}

// SubscriptionsToResponse maps a slice of Subscription entities to SubscriptionResponse DTOs.
func SubscriptionsToResponse(subs []entity.Subscription) []dto.SubscriptionResponse {
	result := make([]dto.SubscriptionResponse, len(subs))
	for i := range subs {
		result[i] = SubscriptionToResponse(&subs[i])
	}
	return result
}
