package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// List returns a paginated list of all videos (admin use).
func (s *Service) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.videoRepo.List(ctx, fs)
	if err != nil {
		s.log.Error("listing videos", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list videos", err)
	}
	return videos, total, nil
}

// ListAdmin returns a paginated list of all videos with per-row view counts,
// optionally restricted to a single genre (admin catalog dashboard).
func (s *Service) ListAdmin(ctx context.Context, fs *filter.FilterSet, genreID *uuid.UUID) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.videoRepo.ListAdmin(ctx, fs, genreID)
	if err != nil {
		s.log.Error("listing admin videos", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list videos", err)
	}
	return videos, total, nil
}

// ListPublished returns a paginated list of published videos (public use).
func (s *Service) ListPublished(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.videoRepo.ListPublished(ctx, fs)
	if err != nil {
		s.log.Error("listing published videos", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list videos", err)
	}
	return videos, total, nil
}

// ListFeatured returns a paginated list of featured published videos (public use).
func (s *Service) ListFeatured(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.videoRepo.ListFeatured(ctx, fs)
	if err != nil {
		s.log.Error("listing featured videos", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list featured videos", err)
	}
	return videos, total, nil
}

// ListBySeries returns a paginated list of episodes belonging to a series.
func (s *Service) ListBySeries(ctx context.Context, seriesID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.videoRepo.ListBySeries(ctx, seriesID, fs)
	if err != nil {
		s.log.Error("listing series episodes", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list episodes", err)
	}
	return videos, total, nil
}
