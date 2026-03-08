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

// UpdateContentInput holds the optional fields for updating content.
// Film-specific fields (BackdropURL, TrailerURL, Director, ReleaseYear) are ignored for Video.
// Video-specific field (CreatorName) is ignored for Film and Series.
type UpdateContentInput struct {
	Title       *string
	Description *string
	Synopsis    *string
	Rating      *string
	ReleaseYear *int
	PosterURL   *string
	BackdropURL *string
	TrailerURL  *string
	Director    *string
	IsFeatured  *bool
	CreatorName *string
}

// Update applies partial updates to existing content, dispatching by type.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateContentInput) (*ContentDetail, *apperror.AppError) {
	ct, err := s.contentRepo.GetTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", err)
		}
		s.log.Error("getting content type for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}

	switch ct {
	case entity.ContentTypeFilm:
		return s.updateFilm(ctx, id, input)
	case entity.ContentTypeVideo:
		return s.updateVideo(ctx, id, input)
	case entity.ContentTypeSeries:
		return s.updateSeries(ctx, id, input)
	default:
		return nil, apperror.Internal(apperror.CodeInternalError, "unknown content type", nil)
	}
}

func (s *Service) updateFilm(ctx context.Context, id uuid.UUID, input UpdateContentInput) (*ContentDetail, *apperror.AppError) {
	film, err := s.contentRepo.GetFilmByID(ctx, id)
	if err != nil || film == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "film not found", domain.ErrContentNotFound)
	}
	s.applyBaseUpdates(ctx, &film.BaseContent, input)
	if input.BackdropURL != nil {
		film.BackdropURL = *input.BackdropURL
	}
	if input.TrailerURL != nil {
		film.TrailerURL = *input.TrailerURL
	}
	if input.Director != nil {
		film.Director = *input.Director
	}
	if input.ReleaseYear != nil {
		film.ReleaseYear = *input.ReleaseYear
	}
	if err := s.contentRepo.UpdateFilm(ctx, film); err != nil {
		s.log.Error("updating film", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update film", err)
	}
	return &ContentDetail{Type: entity.ContentTypeFilm, Film: film}, nil
}

func (s *Service) updateVideo(ctx context.Context, id uuid.UUID, input UpdateContentInput) (*ContentDetail, *apperror.AppError) {
	video, err := s.contentRepo.GetVideoByID(ctx, id)
	if err != nil || video == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "video not found", domain.ErrContentNotFound)
	}
	s.applyBaseUpdates(ctx, &video.BaseContent, input)
	if input.CreatorName != nil {
		video.CreatorName = *input.CreatorName
	}
	if err := s.contentRepo.UpdateVideo(ctx, video); err != nil {
		s.log.Error("updating video", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update video", err)
	}
	return &ContentDetail{Type: entity.ContentTypeVideo, Video: video}, nil
}

func (s *Service) updateSeries(ctx context.Context, id uuid.UUID, input UpdateContentInput) (*ContentDetail, *apperror.AppError) {
	series, err := s.contentRepo.GetSeriesByID(ctx, id)
	if err != nil || series == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "series not found", domain.ErrContentNotFound)
	}
	if input.Title != nil && *input.Title != series.Title {
		series.Title = *input.Title
		baseSlug, slugErr := valueobject.NewSlug(*input.Title)
		if slugErr == nil {
			exists, existsErr := s.contentRepo.ExistsBySlug(ctx, baseSlug.String())
			if existsErr != nil {
				s.log.Error("checking slug existence", "error", existsErr)
			} else if exists {
				series.Slug = baseSlug.WithRandomID()
			} else {
				series.Slug = baseSlug
			}
		}
	}
	if input.Description != nil {
		series.Description = *input.Description
	}
	if input.Synopsis != nil {
		series.Synopsis = *input.Synopsis
	}
	if input.PosterURL != nil {
		series.PosterURL = *input.PosterURL
	}
	if input.BackdropURL != nil {
		series.BackdropURL = *input.BackdropURL
	}
	if input.TrailerURL != nil {
		series.TrailerURL = *input.TrailerURL
	}
	if input.IsFeatured != nil {
		series.IsFeatured = *input.IsFeatured
	}
	series.UpdatedAt = time.Now()
	if err := s.contentRepo.UpdateSeries(ctx, series); err != nil {
		s.log.Error("updating series", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update series", err)
	}
	return &ContentDetail{Type: entity.ContentTypeSeries, Series: series}, nil
}

// applyBaseUpdates applies shared BaseContent fields from the update input.
func (s *Service) applyBaseUpdates(ctx context.Context, base *entity.BaseContent, input UpdateContentInput) {
	if input.Title != nil && *input.Title != base.Title {
		base.Title = *input.Title
		baseSlug, err := valueobject.NewSlug(*input.Title)
		if err == nil {
			exists, existsErr := s.contentRepo.ExistsBySlug(ctx, baseSlug.String())
			if existsErr != nil {
				s.log.Error("checking slug existence", "error", existsErr)
			} else if exists {
				base.Slug = baseSlug.WithRandomID()
			} else {
				base.Slug = baseSlug
			}
		}
	}
	if input.Description != nil {
		base.Description = *input.Description
	}
	if input.Synopsis != nil {
		base.Synopsis = *input.Synopsis
	}
	if input.Rating != nil {
		base.Rating = valueobject.ContentRating(*input.Rating)
	}
	if input.PosterURL != nil {
		base.PosterURL = *input.PosterURL
	}
	if input.IsFeatured != nil {
		base.IsFeatured = *input.IsFeatured
	}
	base.UpdatedAt = time.Now()
}
