package cast

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// ListVideosInput carries the cast ID and pagination parameters.
type ListVideosInput struct {
	CastID  uuid.UUID
	Page    int
	PerPage int
}

// ListVideosOutput holds the paginated list of video credits for a cast member.
type ListVideosOutput struct {
	Credits []entity.VideoCast
	Total   int
	Page    int
	PerPage int
}

// ListVideos returns a paginated list of published videos a cast member appears in.
func (s *Service) ListVideos(ctx context.Context, input ListVideosInput) (*ListVideosOutput, *apperror.AppError) {
	cast, err := s.castRepo.GetByID(ctx, input.CastID)
	if err != nil {
		s.log.Error("getting cast member for video list", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get cast member", err)
	}
	if cast == nil {
		return nil, apperror.NotFound(apperror.CodeCastNotFound, "cast member not found", domain.ErrCastNotFound)
	}

	p := valueobject.NewPagination(input.Page, input.PerPage)

	credits, total, err := s.videoCastRepo.ListByCast(ctx, input.CastID, p.Offset(), p.Limit())
	if err != nil {
		s.log.Error("listing videos by cast", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list videos", err)
	}

	return &ListVideosOutput{
		Credits: credits,
		Total:   total,
		Page:    p.Page,
		PerPage: p.PerPage,
	}, nil
}
