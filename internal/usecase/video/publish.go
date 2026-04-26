package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Publish moves a video into the processing state and triggers HLS transcoding on the CDN.
// The video only becomes visible to viewers once the CDN signals completion via MarkTranscodeComplete.
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

	// Trigger HLS transcoding on the CDN (fire-and-forget).
	// The CDN calls MarkTranscodeComplete / MarkTranscodeFailed via the internal callback when finished.
	if s.cdn != nil {
		go func() {
			if err := s.cdn.TriggerTranscode(video.ID.String()); err != nil {
				s.log.Error("triggering CDN transcode", "video_id", video.ID, "error", err)
			}
		}()
	}

	return video, nil
}
