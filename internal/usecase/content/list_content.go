package content

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// List returns paginated content with optional filtering and search.
func (s *Service) List(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.Content, int64, *apperror.AppError) {
	contents, total, err := s.contentRepo.Search(ctx, fs, searchQuery)
	if err != nil {
		s.log.Error("listing content", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list content", err)
	}
	return contents, total, nil
}

// Search performs a content search (delegates to List).
func (s *Service) Search(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.Content, int64, *apperror.AppError) {
	return s.List(ctx, fs, searchQuery)
}

// AdminList returns paginated content for admin without status filtering.
func (s *Service) AdminList(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.Content, int64, *apperror.AppError) {
	contents, total, err := s.contentRepo.AdminSearch(ctx, fs, searchQuery)
	if err != nil {
		s.log.Error("admin listing content", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list content", err)
	}
	return contents, total, nil
}

// Featured returns featured content up to the given limit.
func (s *Service) Featured(ctx context.Context, limit int) ([]entity.Content, *apperror.AppError) {
	contents, err := s.contentRepo.GetFeatured(ctx, limit)
	if err != nil {
		s.log.Error("getting featured content", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get featured content", err)
	}
	return contents, nil
}
