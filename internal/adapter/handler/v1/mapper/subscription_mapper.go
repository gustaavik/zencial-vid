package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// PlanToResponse maps a Plan entity to a DTO.
func PlanToResponse(p *entity.Plan) dto.PlanResponse {
	return dto.PlanResponse{
		ID:               p.ID.String(),
		Name:             p.Name,
		Tier:             string(p.Tier),
		PriceAmount:      p.Price.Amount,
		PriceCurrency:    p.Price.Currency,
		BillingInterval:  p.BillingInterval,
		MaxQuality:       string(p.MaxQuality),
		MaxStreams:       p.MaxStreams,
		DownloadsAllowed: p.DownloadsAllowed,
	}
}

// PlansToResponse maps a slice of Plan entities to DTOs.
func PlansToResponse(plans []entity.Plan) []dto.PlanResponse {
	result := make([]dto.PlanResponse, len(plans))
	for i := range plans {
		result[i] = PlanToResponse(&plans[i])
	}
	return result
}

// SubscriptionToResponse maps a Subscription entity to a DTO.
func SubscriptionToResponse(s *entity.Subscription) dto.SubscriptionResponse {
	resp := dto.SubscriptionResponse{
		ID:                 s.ID.String(),
		Status:             string(s.Status),
		CurrentPeriodStart: s.CurrentPeriodStart.Format("2006-01-02T15:04:05Z"),
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z"),
		CreatedAt:          s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.Plan != nil {
		resp.Plan = PlanToResponse(s.Plan)
	}
	if s.CanceledAt != nil {
		ca := s.CanceledAt.Format("2006-01-02T15:04:05Z")
		resp.CanceledAt = &ca
	}
	return resp
}
