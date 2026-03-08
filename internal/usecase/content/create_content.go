package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CreateContentInput holds the data required to create new content.
type CreateContentInput struct {
	Type        string
	Title       string
	Description string
	Synopsis    string
	Rating      string
	ReleaseYear int
	PosterURL   string
	BackdropURL string
	TrailerURL  string
	Director    string
	CreatorName string
}

// Create creates new content from the given input, dispatching to the correct typed creator.
func (s *Service) Create(ctx context.Context, input CreateContentInput) (*ContentDetail, *apperror.AppError) {
	slug, appErr := s.uniqueSlug(ctx, input.Title)
	if appErr != nil {
		return nil, appErr
	}

	rating := valueobject.ContentRating(input.Rating)
	if input.Rating == "" && input.Type == string(entity.ContentTypeVideo) {
		rating = valueobject.RatingG
	}

	now := time.Now()
	base := entity.BaseContent{
		ID:          uuid.New(),
		Type:        entity.ContentType(input.Type),
		Title:       input.Title,
		Slug:        slug,
		Description: input.Description,
		Synopsis:    input.Synopsis,
		Rating:      rating,
		PosterURL:   input.PosterURL,
		Status:      entity.ContentStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	switch entity.ContentType(input.Type) {
	case entity.ContentTypeFilm:
		film := &entity.Film{
			BaseContent: base,
			BackdropURL: input.BackdropURL,
			TrailerURL:  input.TrailerURL,
			ReleaseYear: input.ReleaseYear,
			Director:    input.Director,
		}
		if err := s.contentRepo.CreateFilm(ctx, film); err != nil {
			if errors.Is(err, domain.ErrSlugAlreadyExists) {
				return nil, apperror.Conflict(apperror.CodeSlugConflict, "content with this slug already exists", err)
			}
			s.log.Error("creating film", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to create film", err)
		}
		return &ContentDetail{Type: entity.ContentTypeFilm, Film: film}, nil

	case entity.ContentTypeVideo:
		video := &entity.Video{
			BaseContent: base,
			CreatorName: input.CreatorName,
			UploadedAt:  now,
		}
		if err := s.contentRepo.CreateVideo(ctx, video); err != nil {
			if errors.Is(err, domain.ErrSlugAlreadyExists) {
				return nil, apperror.Conflict(apperror.CodeSlugConflict, "content with this slug already exists", err)
			}
			s.log.Error("creating video", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to create video", err)
		}
		return &ContentDetail{Type: entity.ContentTypeVideo, Video: video}, nil

	case entity.ContentTypeSeries:
		series := &entity.Series{
			ID:          base.ID,
			Title:       input.Title,
			Slug:        slug,
			Description: input.Description,
			Synopsis:    input.Synopsis,
			PosterURL:   input.PosterURL,
			BackdropURL: input.BackdropURL,
			TrailerURL:  input.TrailerURL,
			Status:      entity.ContentStatusDraft,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.contentRepo.CreateSeries(ctx, series); err != nil {
			if errors.Is(err, domain.ErrSlugAlreadyExists) {
				return nil, apperror.Conflict(apperror.CodeSlugConflict, "content with this slug already exists", err)
			}
			s.log.Error("creating series", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to create series", err)
		}
		return &ContentDetail{Type: entity.ContentTypeSeries, Series: series}, nil

	default:
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "unknown content type", nil)
	}
}

// uniqueSlug generates a unique slug for the given title.
func (s *Service) uniqueSlug(ctx context.Context, title string) (valueobject.Slug, *apperror.AppError) {
	baseSlug, err := valueobject.NewSlug(title)
	if err != nil {
		return valueobject.Slug{}, apperror.BadRequest(apperror.CodeValidationFailed, "invalid title for slug", err)
	}
	exists, err := s.contentRepo.ExistsBySlug(ctx, baseSlug.String())
	if err != nil {
		s.log.Error("checking slug existence", "error", err)
		return valueobject.Slug{}, apperror.Internal(apperror.CodeInternalError, "failed to check slug", err)
	}
	if exists {
		return baseSlug.WithRandomID(), nil
	}
	return baseSlug, nil
}
