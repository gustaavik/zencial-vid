package content

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Publish transitions content to published status.
func (s *Service) Publish(ctx context.Context, id uuid.UUID) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil || content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
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
