package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetProfile returns the authenticated user's profile.
func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.User, *apperror.AppError) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("getting user profile", "error", err, "userID", userID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get user profile", err)
	}
	if user == nil {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}
	if user.Status == entity.UserStatusDeleted {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserDeleted)
	}

	return user, nil
}
