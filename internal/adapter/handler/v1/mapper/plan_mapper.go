package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// PlanToResponse maps a Plan entity to a PlanResponse DTO.
func PlanToResponse(plan *entity.Plan) dto.PlanResponse {
	return dto.PlanResponse{
		ID:            plan.ID.String(),
		Name:          plan.Name,
		Slug:          plan.Slug.String(),
		Description:   plan.Description,
		Price:         plan.Price,
		Level:         plan.Level,
		StripePriceID: plan.StripePriceID,
		IsActive:      plan.IsActive,
		CreatedAt:     plan.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     plan.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// PlansToResponse maps a slice of Plan entities to PlanResponse DTOs.
func PlansToResponse(plans []entity.Plan) []dto.PlanResponse {
	result := make([]dto.PlanResponse, len(plans))
	for i := range plans {
		result[i] = PlanToResponse(&plans[i])
	}
	return result
}
