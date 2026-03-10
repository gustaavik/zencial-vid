package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetByID returns a video by its ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting video", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}
	return video, nil
}
