package cast

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateCastInput holds the fields needed to update a cast member's name globally.
type UpdateCastInput struct {
	ID   uuid.UUID
	Name *string
	// CallerID and CallerRoles enforce publisher ownership.
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
}

// UpdateCast modifies a cast member's name.
// Admins are unrestricted. Publishers must have the cast member credited on
// at least one of their own videos.
func (s *Service) UpdateCast(ctx context.Context, input *UpdateCastInput) (*entity.Cast, *apperror.AppError) {
	c, err := s.castRepo.GetByID(ctx, input.ID)
	if err != nil {
		s.log.Error("getting cast for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get cast member", err)
	}
	if c == nil {
		return nil, apperror.NotFound(apperror.CodeCastNotFound, "cast member not found", domain.ErrCastNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) {
		ok, err := s.castRepo.HasVideoWithCaller(ctx, input.ID, input.CallerID)
		if err != nil {
			s.log.Error("checking cast ownership for update", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to check ownership", err)
		}
		if !ok {
			return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not have a video with this cast member", domain.ErrVideoOwnershipRequired)
		}
	}

	if input.Name != nil {
		c.Name = *input.Name
	}

	if err := s.castRepo.Update(ctx, c); err != nil {
		s.log.Error("updating cast", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update cast member", err)
	}
	s.resolvePictureURL(c)
	return c, nil
}

// UpdateCreditInput holds the fields needed to update a credit's role/sort_order.
type UpdateCreditInput struct {
	CreditID   uuid.UUID
	Role       *string
	Department *entity.CastDepartment
	SortOrder  *int
	// CallerID and CallerRoles enforce publisher ownership.
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
}

// UpdateCredit modifies the role or sort_order of a cast member's credit on a video.
// Publishers may only update credits for videos they uploaded.
func (s *Service) UpdateCredit(ctx context.Context, input *UpdateCreditInput) (*entity.VideoCast, *apperror.AppError) {
	vc, err := s.videoCastRepo.GetByID(ctx, input.CreditID)
	if err != nil {
		s.log.Error("getting credit for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get cast credit", err)
	}
	if vc == nil {
		return nil, apperror.NotFound(apperror.CodeCastNotFound, "cast credit not found", domain.ErrCastNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) {
		video, err := s.videoRepo.GetByID(ctx, vc.VideoID)
		if err != nil {
			s.log.Error("getting video for credit ownership check", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
		}
		if video == nil || video.UploadedBy != input.CallerID {
			return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
		}
	}

	if input.Role != nil {
		vc.Role = *input.Role
	}
	if input.Department != nil {
		vc.Department = *input.Department
	}
	if input.SortOrder != nil {
		vc.SortOrder = *input.SortOrder
	}

	if err := s.videoCastRepo.Update(ctx, vc); err != nil {
		s.log.Error("updating cast credit", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update cast credit", err)
	}
	s.resolvePictureURL(vc.Cast)
	return vc, nil
}
