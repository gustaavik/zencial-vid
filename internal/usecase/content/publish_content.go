package content

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Publish transitions content to published status.
// For film and video types, a video asset must be attached before publishing.
func (s *Service) Publish(ctx context.Context, id uuid.UUID) (*ContentDetail, *apperror.AppError) {
	ct, err := s.contentRepo.GetTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", err)
		}
		s.log.Error("getting content type for publish", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}

	// Require a video asset for film and video content.
	if ct == entity.ContentTypeFilm || ct == entity.ContentTypeVideo {
		asset, err := s.contentRepo.GetVideoAssetForContent(ctx, id)
		if err != nil {
			s.log.Error("checking video asset for publish", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to check video asset", err)
		}
		if asset == nil {
			return nil, apperror.BadRequest(apperror.CodeVideoAssetRequired, "a video file must be uploaded before publishing", nil)
		}
	}

	if err := s.contentRepo.SetStatus(ctx, id, entity.ContentStatusPublished); err != nil {
		s.log.Error("publishing content", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to publish content", err)
	}
	return s.loadByType(ctx, ct, id)
}

// Archive transitions content to archived status.
func (s *Service) Archive(ctx context.Context, id uuid.UUID) (*ContentDetail, *apperror.AppError) {
	ct, err := s.contentRepo.GetTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", err)
		}
		s.log.Error("getting content type for archive", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}

	if err := s.contentRepo.SetStatus(ctx, id, entity.ContentStatusArchived); err != nil {
		s.log.Error("archiving content", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to archive content", err)
	}
	return s.loadByType(ctx, ct, id)
}
