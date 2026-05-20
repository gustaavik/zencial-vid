package cast

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateInput holds the fields needed to update a cast member.
type UpdateInput struct {
	ID        uuid.UUID
	Name      *string
	Role      *string
	SortOrder *int
	// CallerID and CallerRoles enforce publisher ownership.
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
}

// Update modifies an existing cast member.
// Publishers may only update cast for videos they uploaded.
func (s *Service) Update(ctx context.Context, input *UpdateInput) (*entity.Cast, *apperror.AppError) {
	c, err := s.castRepo.GetByID(ctx, input.ID)
	if err != nil {
		s.log.Error("getting cast for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get cast member", err)
	}
	if c == nil {
		return nil, apperror.NotFound(apperror.CodeCastNotFound, "cast member not found", domain.ErrCastNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) {
		video, err := s.videoRepo.GetByID(ctx, c.VideoID)
		if err != nil {
			s.log.Error("getting video for cast ownership check", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
		}
		if video == nil || video.UploadedBy != input.CallerID {
			return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
		}
	}

	if input.Name != nil {
		c.Name = *input.Name
	}
	if input.Role != nil {
		c.Role = *input.Role
	}
	if input.SortOrder != nil {
		c.SortOrder = *input.SortOrder
	}

	if err := s.castRepo.Update(ctx, c); err != nil {
		s.log.Error("updating cast", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update cast member", err)
	}
	return c, nil
}
