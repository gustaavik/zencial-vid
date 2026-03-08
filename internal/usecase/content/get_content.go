package content

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetByID retrieves typed content by its ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*ContentDetail, *apperror.AppError) {
	ct, err := s.contentRepo.GetTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", err)
		}
		s.log.Error("getting content type by id", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	return s.loadByType(ctx, ct, id)
}

// GetBySlug retrieves typed content by its URL slug.
// It tries each typed getter in order (film → video → series).
func (s *Service) GetBySlug(ctx context.Context, slug string) (*ContentDetail, *apperror.AppError) {
	film, err := s.contentRepo.GetFilmBySlug(ctx, slug)
	if err != nil {
		s.log.Error("getting film by slug", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if film != nil {
		return &ContentDetail{Type: entity.ContentTypeFilm, Film: film}, nil
	}

	video, err := s.contentRepo.GetVideoBySlug(ctx, slug)
	if err != nil {
		s.log.Error("getting video by slug", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if video != nil {
		return &ContentDetail{Type: entity.ContentTypeVideo, Video: video}, nil
	}

	series, err := s.contentRepo.GetSeriesBySlug(ctx, slug)
	if err != nil {
		s.log.Error("getting series by slug", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if series != nil {
		return &ContentDetail{Type: entity.ContentTypeSeries, Series: series}, nil
	}

	return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
}

// loadByType fetches the full typed entity for a given ID and content type.
func (s *Service) loadByType(ctx context.Context, ct entity.ContentType, id uuid.UUID) (*ContentDetail, *apperror.AppError) {
	switch ct {
	case entity.ContentTypeFilm:
		film, err := s.contentRepo.GetFilmByID(ctx, id)
		if err != nil || film == nil {
			return nil, apperror.NotFound(apperror.CodeContentNotFound, "film not found", domain.ErrContentNotFound)
		}
		return &ContentDetail{Type: ct, Film: film}, nil

	case entity.ContentTypeVideo:
		video, err := s.contentRepo.GetVideoByID(ctx, id)
		if err != nil || video == nil {
			return nil, apperror.NotFound(apperror.CodeContentNotFound, "video not found", domain.ErrContentNotFound)
		}
		return &ContentDetail{Type: ct, Video: video}, nil

	case entity.ContentTypeSeries:
		series, err := s.contentRepo.GetSeriesByID(ctx, id)
		if err != nil || series == nil {
			return nil, apperror.NotFound(apperror.CodeContentNotFound, "series not found", domain.ErrContentNotFound)
		}
		return &ContentDetail{Type: ct, Series: series}, nil

	default:
		return nil, apperror.Internal(apperror.CodeInternalError, "unknown content type", nil)
	}
}
