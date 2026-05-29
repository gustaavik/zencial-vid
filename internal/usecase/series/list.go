package series

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// List returns all series (admin use).
func (s *Service) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, *apperror.AppError) {
	series, total, err := s.seriesRepo.List(ctx, fs)
	if err != nil {
		s.log.Error("listing series", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list series", err)
	}
	return series, total, nil
}

// ListPublished returns published series for the public feed.
func (s *Service) ListPublished(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, *apperror.AppError) {
	series, total, err := s.seriesRepo.ListPublished(ctx, fs)
	if err != nil {
		s.log.Error("listing published series", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list series", err)
	}
	return series, total, nil
}

// ListOwned returns series uploaded by a specific publisher.
func (s *Service) ListOwned(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Series, int64, *apperror.AppError) {
	series, total, err := s.seriesRepo.ListByUploader(ctx, uploaderID, fs)
	if err != nil {
		s.log.Error("listing owned series", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list series", err)
	}
	return series, total, nil
}
