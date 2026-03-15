package genre

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CreateInput holds the data needed to create a genre.
type CreateInput struct {
	Slug         string
	Translations []TranslationInput
}

// TranslationInput holds the data for a genre translation.
type TranslationInput struct {
	LanguageCode string
	Name         string
	Description  string
}

// Create creates a new genre with translations.
func (s *Service) Create(ctx context.Context, input CreateInput) (*entity.Genre, *apperror.AppError) {
	slug, err := valueobject.NewSlug(input.Slug)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid slug", err)
	}

	exists, err := s.genreRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		s.log.Error("checking slug existence", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check slug", err)
	}
	if exists {
		return nil, apperror.Conflict(apperror.CodeSlugConflict, "slug already exists", domain.ErrSlugAlreadyExists)
	}

	genre := entity.NewGenre(slug)
	for _, t := range input.Translations {
		langCode, err := valueobject.NewLanguageCode(t.LanguageCode)
		if err != nil {
			return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid language code", err)
		}
		genre.AddTranslation(langCode, t.Name, t.Description)
	}

	if err := s.genreRepo.Create(ctx, genre); err != nil {
		s.log.Error("creating genre", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create genre", err)
	}

	return genre, nil
}
