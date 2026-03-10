package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Delete removes a video and its associated storage files.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting video for delete", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	// Delete from storage (best effort)
	if err := s.storage.Delete(ctx, video.StorageKey); err != nil {
		s.log.Error("deleting video from storage", "error", err, "key", video.StorageKey)
	}
	if video.ThumbnailKey != "" {
		if err := s.storage.Delete(ctx, video.ThumbnailKey); err != nil {
			s.log.Error("deleting thumbnail from storage", "error", err, "key", video.ThumbnailKey)
		}
	}

	if err := s.videoRepo.Delete(ctx, id); err != nil {
		s.log.Error("deleting video record", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete video", err)
	}

	return nil
}
