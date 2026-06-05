package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// SetFeatured marks a published video as featured on the home page.
// Returns an error if the video does not exist or is not published.
func (s *Service) SetFeatured(ctx context.Context, videoID uuid.UUID, description string) *apperror.AppError {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("fetching video for set-featured", "video_id", videoID, "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}
	if !video.IsPlayable() {
		return apperror.Conflict(apperror.CodeVideoNotPublished, "only published videos can be featured", domain.ErrVideoNotPublished)
	}

	if err := s.videoRepo.SetFeatured(ctx, videoID, description); err != nil {
		s.log.Error("setting video as featured", "video_id", videoID, "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to set video as featured", err)
	}
	return nil
}

// UnsetFeatured removes a video from the featured list.
func (s *Service) UnsetFeatured(ctx context.Context, videoID uuid.UUID) *apperror.AppError {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("fetching video for unset-featured", "video_id", videoID, "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if err := s.videoRepo.UnsetFeatured(ctx, videoID); err != nil {
		s.log.Error("unsetting video as featured", "video_id", videoID, "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to unset video as featured", err)
	}
	return nil
}

