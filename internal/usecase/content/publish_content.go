package content

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Publish transitions content to published status.
// For film and video types, a video asset must be attached before publishing.
func (s *Service) Publish(ctx context.Context, id uuid.UUID) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil || content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}

	// Require a video asset for film and video content.
	if content.Type == entity.ContentTypeFilm || content.Type == entity.ContentTypeVideo {
		asset, err := s.contentRepo.GetVideoAssetForContent(ctx, content.ID)
		if err != nil {
			s.log.Error("checking video asset for publish", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to check video asset", err)
		}
		if asset == nil {
			return nil, apperror.BadRequest(apperror.CodeVideoAssetRequired, "a video file must be uploaded before publishing", nil)
		}
	}

	content.Publish()
	if err := s.contentRepo.Update(ctx, content); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to publish content", err)
	}
	return content, nil
}

// Archive transitions content to archived status.
func (s *Service) Archive(ctx context.Context, id uuid.UUID) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil || content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}
	content.Archive()
	if err := s.contentRepo.Update(ctx, content); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to archive content", err)
	}
	return content, nil
}
