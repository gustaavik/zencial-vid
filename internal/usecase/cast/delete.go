package cast

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Delete removes a cast member.
// Publishers may only delete cast for videos they uploaded.
func (s *Service) Delete(ctx context.Context, id, callerID uuid.UUID, callerRole entity.UserRole) *apperror.AppError {
	c, err := s.castRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting cast for delete", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get cast member", err)
	}
	if c == nil {
		return apperror.NotFound(apperror.CodeCastNotFound, "cast member not found", domain.ErrCastNotFound)
	}

	if callerRole == entity.RolePublisher {
		video, err := s.videoRepo.GetByID(ctx, c.VideoID)
		if err != nil {
			s.log.Error("getting video for cast ownership check", "error", err)
			return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
		}
		if video == nil || video.UploadedBy != callerID {
			return apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
		}
	}

	if err := s.castRepo.Delete(ctx, id); err != nil {
		s.log.Error("deleting cast", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete cast member", err)
	}
	return nil
}
