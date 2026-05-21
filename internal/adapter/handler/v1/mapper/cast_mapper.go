package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// VideoCastToResponse converts a VideoCast entity to a CastCreditResponse DTO.
func VideoCastToResponse(vc *entity.VideoCast) dto.CastCreditResponse {
	r := dto.CastCreditResponse{
		ID:        vc.CastID.String(),
		VideoID:   vc.VideoID.String(),
		Role:      vc.Role,
		SortOrder: vc.SortOrder,
		CreatedAt: vc.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: vc.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if vc.Cast != nil {
		r.Name = vc.Cast.Name
		r.PictureURL = vc.Cast.PictureURL
	}
	return r
}

// VideoCastListToResponse converts a slice of VideoCast entities to CastCreditResponse DTOs.
func VideoCastListToResponse(credits []entity.VideoCast) []dto.CastCreditResponse {
	out := make([]dto.CastCreditResponse, len(credits))
	for i := range credits {
		out[i] = VideoCastToResponse(&credits[i])
	}
	return out
}

// CastMembersToResponse converts a slice of Cast entities to CastMemberResponse DTOs.
func CastMembersToResponse(casts []entity.Cast) []dto.CastMemberResponse {
	out := make([]dto.CastMemberResponse, len(casts))
	for i := range casts {
		out[i] = CastToMemberResponse(&casts[i])
	}
	return out
}

// CastToMemberResponse converts a Cast entity to a CastMemberResponse DTO.
func CastToMemberResponse(c *entity.Cast) dto.CastMemberResponse {
	return dto.CastMemberResponse{
		ID:         c.ID.String(),
		Name:       c.Name,
		PictureURL: c.PictureURL,
		CreatedAt:  c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
