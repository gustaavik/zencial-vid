package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
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
