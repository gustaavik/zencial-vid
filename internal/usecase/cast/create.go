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
	VideoID    uuid.UUID
	Name       string
	Role       string
	Department entity.CastDepartment
	SortOrder  int
	// CallerID and CallerRoles are used to enforce publisher ownership.
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
}

// Create adds a cast member to a video using find-or-create by name.
// Publishers may only add cast to videos they uploaded; admins are unrestricted.
func (s *Service) Create(ctx context.Context, input *CreateInput) (*entity.VideoCast, *apperror.AppError) {
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

	cast, err := s.castRepo.FindOrCreate(ctx, input.Name)
	if err != nil {
		s.log.Error("finding or creating cast member", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to resolve cast member", err)
	}

	if cast.IsArchived() {
		return nil, apperror.Conflict(apperror.CodeCastArchived, "cast member is archived; unarchive before crediting", domain.ErrCastArchived)
	}

	existing, err := s.videoCastRepo.GetByVideoAndCastAndRole(ctx, input.VideoID, cast.ID, input.Role)
	if err != nil {
		s.log.Error("checking for duplicate cast credit", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check cast credit", err)
	}
	if existing != nil {
		return nil, apperror.Conflict(apperror.CodeCastAlreadyCredited, "cast member already has this role on this video", nil)
	}

	dept := input.Department
	if dept == "" {
		dept = entity.DepartmentPerformance
	}
	vc := entity.NewVideoCast(input.VideoID, cast.ID, input.Role, dept, input.SortOrder)
	vc.Cast = cast
	if err := s.videoCastRepo.Create(ctx, vc); err != nil {
		s.log.Error("creating cast credit", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create cast credit", err)
	}
	s.resolvePictureURL(vc.Cast)
	return vc, nil
}
