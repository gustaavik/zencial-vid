package genre

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// List returns a paginated list of genres.
func (s *Service) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Genre, int64, *apperror.AppError) {
	genres, total, err := s.genreRepo.List(ctx, fs)
	if err != nil {
		s.log.Error("listing genres", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list genres", err)
	}
	return genres, total, nil
}
