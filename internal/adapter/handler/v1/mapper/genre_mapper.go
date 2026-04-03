package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	genreuc "github.com/zenfulcode/zencial/internal/usecase/genre"
)

// GenreToResponse maps a Genre entity to a GenreResponse DTO.
func GenreToResponse(genre *entity.Genre) dto.GenreResponse {
	translations := make([]dto.GenreTranslationResponse, len(genre.Translations))
	for i, t := range genre.Translations {
		translations[i] = dto.GenreTranslationResponse{
			LanguageCode: t.LanguageCode.String(),
			Name:         t.Name,
			Description:  t.Description,
		}
	}
	return dto.GenreResponse{
		ID:           genre.ID.String(),
		Slug:         genre.Slug.String(),
		Translations: translations,
		CreatedAt:    genre.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    genre.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// GenresToResponse maps a slice of Genre entities to GenreResponse DTOs.
func GenresToResponse(genres []entity.Genre) []dto.GenreResponse {
	result := make([]dto.GenreResponse, len(genres))
	for i := range genres {
		result[i] = GenreToResponse(&genres[i])
	}
	return result
}

// BulkCreateGenreResultToResponse maps a BulkCreateResult to a BulkCreateGenreResultResponse DTO.
func BulkCreateGenreResultToResponse(result *genreuc.BulkCreateResult) dto.BulkCreateGenreResultResponse {
	succeeded := make([]dto.GenreResponse, len(result.Succeeded))
	for i, g := range result.Succeeded {
		succeeded[i] = GenreToResponse(g)
	}

	failed := make([]dto.BulkCreateGenreFailureResponse, len(result.Failed))
	for i, f := range result.Failed {
		failed[i] = dto.BulkCreateGenreFailureResponse{
			Slug:  f.Slug,
			Error: f.Error,
		}
	}

	return dto.BulkCreateGenreResultResponse{
		Succeeded: succeeded,
		Failed:    failed,
	}
}

// BulkDeleteGenreResultToResponse maps a BulkDeleteResult to a BulkDeleteGenreResultResponse DTO.
func BulkDeleteGenreResultToResponse(result *genreuc.BulkDeleteResult) dto.BulkDeleteGenreResultResponse {
	succeeded := make([]string, len(result.Succeeded))
	for i, id := range result.Succeeded {
		succeeded[i] = id.String()
	}

	failed := make([]dto.BulkDeleteGenreFailureResponse, len(result.Failed))
	for i, f := range result.Failed {
		failed[i] = dto.BulkDeleteGenreFailureResponse{
			ID:    f.ID.String(),
			Error: f.Error,
		}
	}

	return dto.BulkDeleteGenreResultResponse{
		Succeeded: succeeded,
		Failed:    failed,
	}
}
