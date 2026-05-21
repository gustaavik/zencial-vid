package cast

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// List returns all cast members for a video.
func (s *Service) List(ctx context.Context, videoID uuid.UUID) ([]entity.Cast, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("getting video for cast list", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	cast, err := s.castRepo.ListByVideo(ctx, videoID)
	if err != nil {
		s.log.Error("listing cast", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list cast", err)
	}
	for i := range cast {
		s.resolvePictureURL(&cast[i])
	}
	return cast, nil
}
