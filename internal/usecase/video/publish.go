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

// Publish sets a video's status to published.
func (s *Service) Publish(ctx context.Context, id uuid.UUID) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting video for publish", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	video.Publish()

	if err := s.videoRepo.Update(ctx, video); err != nil {
		s.log.Error("publishing video", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to publish video", err)
	}

	s.dispatcher.Dispatch(event.VideoPublished{
		VideoID:   video.ID,
		Timestamp: time.Now().UTC(),
	})

	return video, nil
}
