package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// CastToResponse converts a Cast entity to a CastResponse DTO.
func CastToResponse(c *entity.Cast) dto.CastResponse {
	return dto.CastResponse{
		ID:        c.ID.String(),
		VideoID:   c.VideoID.String(),
		Name:      c.Name,
		Role:      c.Role,
		SortOrder: c.SortOrder,
		CreatedAt: c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// CastListToResponse converts a slice of Cast entities to CastResponse DTOs.
func CastListToResponse(cast []entity.Cast) []dto.CastResponse {
	out := make([]dto.CastResponse, len(cast))
	for i := range cast {
		out[i] = CastToResponse(&cast[i])
	}
	return out
}
