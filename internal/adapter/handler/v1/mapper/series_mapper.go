package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// SeriesToResponse maps a Series entity to a SeriesResponse DTO.
func SeriesToResponse(series *entity.Series) dto.SeriesResponse {
	genreIDs := make([]string, len(series.GenreIDs))
	for i, gid := range series.GenreIDs {
		genreIDs[i] = gid.String()
	}

	return dto.SeriesResponse{
		ID:               series.ID.String(),
		Title:            series.Title,
		Slug:             series.Slug.String(),
		Description:      series.Description,
		Creator:          series.Creator,
		Status:           string(series.Status),
		CoverImageKey:    series.CoverImageKey,
		GenreIDs:         genreIDs,
		MinimumPlanLevel: series.MinimumPlanLevel,
		CreatedAt:        series.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        series.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// SeriesToResponseMany maps a slice of Series entities to SeriesResponse DTOs.
func SeriesToResponseMany(series []entity.Series) []dto.SeriesResponse {
	result := make([]dto.SeriesResponse, len(series))
	for i := range series {
		result[i] = SeriesToResponse(&series[i])
	}
	return result
}

// SeriesWatchProgressToResponse maps a SeriesWatchProgress entity to a SeriesWatchProgressResponse DTO.
func SeriesWatchProgressToResponse(p *entity.SeriesWatchProgress) dto.SeriesWatchProgressResponse {
	return dto.SeriesWatchProgressResponse{
		SeriesID:      p.SeriesID.String(),
		LastEpisodeID: p.LastEpisodeID.String(),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
