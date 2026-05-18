package analytics

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetVideoStats returns viewing statistics for a single video.
// Publishers may only query stats for videos they uploaded; admins may query any video.
func (s *Service) GetVideoStats(ctx context.Context, videoID, callerID uuid.UUID, callerRoles []entity.UserRole) (*repository.VideoStats, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("getting video for analytics", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if !entity.HasRole(callerRoles, entity.RoleAdmin) && video.UploadedBy != callerID {
		return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
	}

	stats, err := s.analyticsRepo.GetVideoStats(ctx, videoID)
	if err != nil {
		s.log.Error("getting video stats", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video stats", err)
	}
	return stats, nil
}
