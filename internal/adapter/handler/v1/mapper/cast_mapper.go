package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// VideoCastToResponse converts a VideoCast entity to a CastCreditResponse DTO.
func VideoCastToResponse(vc *entity.VideoCast) dto.CastCreditResponse {
	r := dto.CastCreditResponse{
		ID:        vc.ID.String(),
		CastID:    vc.CastID.String(),
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
		Status:     string(c.Status),
		PictureURL: c.PictureURL,
		CreatedAt:  c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// VideoCastToVideoResponse maps a VideoCast (with populated Video) to a CastVideoResponse DTO.
func VideoCastToVideoResponse(vc *entity.VideoCast, urls ThumbnailURLBuilder) dto.CastVideoResponse {
	r := dto.CastVideoResponse{
		VideoID:   vc.VideoID.String(),
		Role:      vc.Role,
		SortOrder: vc.SortOrder,
		CreatedAt: vc.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: vc.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if vc.Video != nil {
		r.Title = vc.Video.Title
		r.Slug = vc.Video.Slug.String()
		r.Status = string(vc.Video.Status)
		r.Duration = vc.Video.Duration.Seconds
		r.ContentRating = vc.Video.ContentRating
		r.SeasonNumber = vc.Video.SeasonNumber
		r.EpisodeNumber = vc.Video.EpisodeNumber
		if vc.Video.SeriesID != nil {
			s := vc.Video.SeriesID.String()
			r.SeriesID = &s
		}
		if vc.Video.ThumbnailKey != "" && urls != nil {
			r.ThumbnailURL = urls.ThumbnailURL(vc.VideoID.String())
		}
	}
	return r
}

// VideoCastToVideoResponses maps a slice of VideoCast entities to CastVideoResponse DTOs.
func VideoCastToVideoResponses(credits []entity.VideoCast, urls ThumbnailURLBuilder) []dto.CastVideoResponse {
	out := make([]dto.CastVideoResponse, len(credits))
	for i := range credits {
		out[i] = VideoCastToVideoResponse(&credits[i], urls)
	}
	return out
}
