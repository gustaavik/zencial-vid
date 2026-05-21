package cast

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CreateInput holds the fields needed to add a cast member to a video.
type CreateInput struct {
	VideoID   uuid.UUID
	Name      string
	Role      string
	SortOrder int
	// CallerID and CallerRoles are used to enforce publisher ownership.
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
}

// Create adds a new cast member to a video.
// Publishers may only add cast to videos they uploaded; admins are unrestricted.
func (s *Service) Create(ctx context.Context, input *CreateInput) (*entity.Cast, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, input.VideoID)
	if err != nil {
		s.log.Error("getting video for cast create", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) && video.UploadedBy != input.CallerID {
		return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
	}

	c := entity.NewCast(input.VideoID, input.Name, input.Role, input.SortOrder)
	if err := s.castRepo.Create(ctx, c); err != nil {
		s.log.Error("creating cast", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create cast member", err)
	}
	s.resolvePictureURL(c)
	return c, nil
}
