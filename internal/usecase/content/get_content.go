package content

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetByID retrieves content by its ID, including video-specific data and assets (no status filter).
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting content by id", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}

	return s.loadContentExtras(ctx, content)
}

// GetBySlug retrieves content by its URL slug, including video-specific data and assets.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetBySlug(ctx, slug)
	if err != nil {
		s.log.Error("getting content by slug", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}

	return s.loadContentExtras(ctx, content)
}

// loadContentExtras loads video-specific data and assets for the given content.
func (s *Service) loadContentExtras(ctx context.Context, content *entity.Content) (*entity.Content, *apperror.AppError) {
	// Load video-specific data.
	if content.Type == entity.ContentTypeVideo {
		video, err := s.contentRepo.GetVideoForContent(ctx, content.ID)
		if err != nil {
			s.log.Error("getting video data", "error", err)
		} else {
			content.Video = video
		}
	}

	// Load video asset for film or video content.
	if content.Type == entity.ContentTypeFilm || content.Type == entity.ContentTypeVideo {
		asset, err := s.contentRepo.GetVideoAssetForContent(ctx, content.ID)
		if err != nil {
			s.log.Error("getting video asset", "error", err)
		} else if asset != nil {
			if content.Type == entity.ContentTypeFilm {
				if content.Film == nil {
					content.Film = &entity.Film{ContentID: content.ID}
				}
				content.Film.Asset = *asset
			} else if content.Video != nil {
				content.Video.Asset = *asset
			}
		}
	}

	return content, nil
}
