package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetByID returns a user by ID (admin operation, includes all statuses).
func (s *Service) GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, *apperror.AppError) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("getting user by ID", "error", err, "userID", userID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get user", err)
	}
	if user == nil {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}

	return user, nil
}
