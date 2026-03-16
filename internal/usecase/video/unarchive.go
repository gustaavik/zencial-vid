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

// Unarchive restores a soft-deleted video by moving its storage files back
// from the deleted/ prefix and setting the status to draft.
func (s *Service) Unarchive(ctx context.Context, id uuid.UUID) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting video for unarchive", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if video.Status != entity.VideoStatusArchived {
		return video, nil
	}

	// Move video file back from deleted/ prefix.
	restoredKey := entity.RestoredStorageKey(video.StorageKey)
	if err := s.storage.Move(ctx, video.StorageKey, restoredKey); err != nil {
		s.log.Error("moving video from deleted prefix", "error", err, "key", video.StorageKey)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to restore video file", err)
	}
	video.StorageKey = restoredKey

	// Move thumbnail back (best effort).
	if video.ThumbnailKey != "" {
		restoredThumb := entity.RestoredStorageKey(video.ThumbnailKey)
		if err := s.storage.Move(ctx, video.ThumbnailKey, restoredThumb); err != nil {
			s.log.Error("moving thumbnail from deleted prefix", "error", err, "key", video.ThumbnailKey)
		} else {
			video.ThumbnailKey = restoredThumb
		}
	}

	video.Unarchive()

	if err := s.videoRepo.Update(ctx, video); err != nil {
		s.log.Error("updating video record for unarchive", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to restore video", err)
	}

	if err := s.dispatcher.Dispatch(event.VideoRestored{
		VideoID:   video.ID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching video restored event", "error", err)
	}

	return video, nil
}
