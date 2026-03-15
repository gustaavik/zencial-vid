package genre

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateInput holds the data needed to update a genre.
type UpdateInput struct {
	ID           uuid.UUID
	Slug         *string
	Translations []TranslationInput
}

// Update updates an existing genre.
func (s *Service) Update(ctx context.Context, input UpdateInput) (*entity.Genre, *apperror.AppError) {
	genre, err := s.genreRepo.GetByID(ctx, input.ID)
	if err != nil {
		s.log.Error("getting genre for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get genre", err)
	}
	if genre == nil {
		return nil, apperror.NotFound(apperror.CodeGenreNotFound, "genre not found", domain.ErrGenreNotFound)
	}

	if input.Slug != nil {
		newSlug, err := valueobject.NewSlug(*input.Slug)
		if err != nil {
			return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid slug", err)
		}
		if newSlug.String() != genre.Slug.String() {
			exists, err := s.genreRepo.ExistsBySlug(ctx, newSlug)
			if err != nil {
				s.log.Error("checking slug existence", "error", err)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to check slug", err)
			}
			if exists {
				return nil, apperror.Conflict(apperror.CodeSlugConflict, "slug already exists", domain.ErrSlugAlreadyExists)
			}
			genre.Slug = newSlug
		}
	}

	// Replace translations entirely
	genre.Translations = nil
	for _, t := range input.Translations {
		langCode, err := valueobject.NewLanguageCode(t.LanguageCode)
		if err != nil {
			return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid language code", err)
		}
		genre.AddTranslation(langCode, t.Name, t.Description)
	}

	if err := s.genreRepo.Update(ctx, genre); err != nil {
		s.log.Error("updating genre", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update genre", err)
	}

	return genre, nil
}
