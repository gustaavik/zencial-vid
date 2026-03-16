package user

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// List returns a paginated and filtered list of users (admin operation).
func (s *Service) List(ctx context.Context, fs *filter.FilterSet) ([]entity.User, int64, *apperror.AppError) {
	users, total, err := s.userRepo.List(ctx, fs)
	if err != nil {
		s.log.Error("listing users", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list users", err)
	}

	return users, total, nil
}
