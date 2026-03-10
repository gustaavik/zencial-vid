package video

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// List returns a paginated list of all videos (admin use).
func (s *Service) List(ctx context.Context, fs filter.FilterSet) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.videoRepo.List(ctx, fs)
	if err != nil {
		s.log.Error("listing videos", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list videos", err)
	}
	return videos, total, nil
}

// ListPublished returns a paginated list of published videos (public use).
func (s *Service) ListPublished(ctx context.Context, fs filter.FilterSet) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.videoRepo.ListPublished(ctx, fs)
	if err != nil {
		s.log.Error("listing published videos", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list videos", err)
	}
	return videos, total, nil
}
