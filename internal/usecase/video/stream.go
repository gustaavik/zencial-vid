package video

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

const defaultStreamURLExpiry = 4 * time.Hour

// StreamOutput holds the presigned streaming URL.
type StreamOutput struct {
	URL       string
	ExpiresAt time.Time
}

// GetStreamURL generates a presigned URL for streaming a published video.
func (s *Service) GetStreamURL(ctx context.Context, videoID uuid.UUID, userID uuid.UUID) (*StreamOutput, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("getting video for streaming", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if !video.IsPlayable() {
		return nil, apperror.BadRequest(apperror.CodeVideoNotPlayable, "video is not available for streaming", domain.ErrVideoNotPlayable)
	}

	// Check subscription if video requires a plan
	if video.RequiresSubscription() {
		sub, err := s.subRepo.GetActiveByUserID(ctx, userID)
		if err != nil {
			s.log.Error("checking subscription for streaming", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to check subscription", err)
		}
		if sub == nil {
			return nil, apperror.Forbidden(apperror.CodeInsufficientPlanLevel, "subscription required to watch this content", domain.ErrInsufficientPlanLevel)
		}

		plan, err := s.planRepo.GetByID(ctx, sub.PlanID)
		if err != nil {
			s.log.Error("getting plan for subscription check", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to check plan", err)
		}
		if plan == nil || plan.Level < *video.MinimumPlanLevel {
			return nil, apperror.Forbidden(apperror.CodeInsufficientPlanLevel, "your plan does not include this content", domain.ErrInsufficientPlanLevel)
		}
	}

	presignedURL, err := s.storage.PresignedGetURL(ctx, video.StorageKey, defaultStreamURLExpiry)
	if err != nil {
		s.log.Error("generating presigned URL", "error", err)
		return nil, apperror.Internal(apperror.CodeStorageError, "failed to generate stream URL", err)
	}

	return &StreamOutput{
		URL:       presignedURL,
		ExpiresAt: time.Now().UTC().Add(defaultStreamURLExpiry),
	}, nil
}
