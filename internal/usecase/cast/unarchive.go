package cast

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UnarchiveInput holds parameters for restoring an archived cast member.
type UnarchiveInput struct {
	ID          uuid.UUID
	CallerRoles []entity.UserRole
}

// UnarchiveOutput holds the restored cast member.
type UnarchiveOutput struct {
	Cast *entity.Cast
}

// Unarchive restores an archived cast member to active status (admin only).
func (s *Service) Unarchive(ctx context.Context, input *UnarchiveInput) (*UnarchiveOutput, *apperror.AppError) {
	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) {
		return nil, apperror.Forbidden(apperror.CodeForbidden, "only admins can unarchive cast members", nil)
	}

	c, err := s.castRepo.GetByID(ctx, input.ID)
	if err != nil {
		s.log.Error("getting cast for unarchive", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get cast member", err)
	}
	if c == nil {
		return nil, apperror.NotFound(apperror.CodeCastNotFound, "cast member not found", domain.ErrCastNotFound)
	}

	if !c.IsArchived() {
		return nil, apperror.BadRequest(apperror.CodeBadRequest, "cast member is not archived", nil)
	}

	c.Unarchive()
	if err := s.castRepo.Update(ctx, c); err != nil {
		s.log.Error("unarchiving cast member", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to unarchive cast member", err)
	}

	s.resolvePictureURL(c)
	return &UnarchiveOutput{Cast: c}, nil
}
