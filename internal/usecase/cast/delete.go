package cast

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// DeleteFromVideo removes a cast member's credit from a single video.
// Publishers may only delete credits for videos they uploaded.
// The cast member record in the casts table is not affected.
func (s *Service) DeleteFromVideo(ctx context.Context, castID, videoID, callerID uuid.UUID, callerRoles []entity.UserRole) *apperror.AppError {
	vc, err := s.videoCastRepo.GetByVideoAndCast(ctx, videoID, castID)
	if err != nil {
		s.log.Error("getting credit for delete", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get cast credit", err)
	}
	if vc == nil {
		return apperror.NotFound(apperror.CodeCastNotFound, "cast credit not found", domain.ErrCastNotFound)
	}

	if !entity.HasRole(callerRoles, entity.RoleAdmin) {
		video, err := s.videoRepo.GetByID(ctx, videoID)
		if err != nil {
			s.log.Error("getting video for credit ownership check", "error", err)
			return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
		}
		if video == nil || video.UploadedBy != callerID {
			return apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
		}
	}

	if err := s.videoCastRepo.DeleteByVideoAndCast(ctx, videoID, castID); err != nil {
		s.log.Error("deleting cast credit", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete cast credit", err)
	}
	return nil
}

// Delete archives a cast member globally (soft delete, admin only).
// The record and all video credits are preserved; the cast member is hidden
// from normal listings until unarchived.
func (s *Service) Delete(ctx context.Context, castID, callerID uuid.UUID, callerRoles []entity.UserRole) *apperror.AppError {
	if !entity.HasRole(callerRoles, entity.RoleAdmin) {
		return apperror.Forbidden(apperror.CodeForbidden, "only admins can archive cast members", nil)
	}

	c, err := s.castRepo.GetByID(ctx, castID)
	if err != nil {
		s.log.Error("getting cast for archive", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get cast member", err)
	}
	if c == nil {
		return apperror.NotFound(apperror.CodeCastNotFound, "cast member not found", domain.ErrCastNotFound)
	}

	if c.IsArchived() {
		return nil
	}

	c.Archive()
	if err := s.castRepo.Update(ctx, c); err != nil {
		s.log.Error("archiving cast member", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to archive cast member", err)
	}
	return nil
}
