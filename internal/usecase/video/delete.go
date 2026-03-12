package video

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Delete soft-deletes a video by moving its storage files to a deleted/ prefix
// and marking the video as archived.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting video for delete", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if video.Status == entity.VideoStatusArchived {
		return nil
	}

	// Move video file to deleted/ prefix.
	newStorageKey := entity.DeletedStorageKey(video.StorageKey)
	if err := s.storage.Move(ctx, video.StorageKey, newStorageKey); err != nil {
		s.log.Error("moving video to deleted prefix", "error", err, "key", video.StorageKey)
		return apperror.Internal(apperror.CodeInternalError, "failed to archive video file", err)
	}
	video.StorageKey = newStorageKey

	// Move thumbnail to deleted/ prefix (best effort).
	if video.ThumbnailKey != "" {
		newThumbnailKey := entity.DeletedStorageKey(video.ThumbnailKey)
		if err := s.storage.Move(ctx, video.ThumbnailKey, newThumbnailKey); err != nil {
			s.log.Error("moving thumbnail to deleted prefix", "error", err, "key", video.ThumbnailKey)
		} else {
			video.ThumbnailKey = newThumbnailKey
		}
	}

	video.Archive()

	if err := s.videoRepo.Update(ctx, video); err != nil {
		s.log.Error("updating video record for soft delete", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to archive video", err)
	}

	s.dispatcher.Dispatch(event.VideoArchived{
		VideoID:   video.ID,
		Timestamp: time.Now().UTC(),
	})

	return nil
}
