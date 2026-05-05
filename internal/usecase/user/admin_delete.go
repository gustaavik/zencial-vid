package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminDelete soft-deletes a user account on behalf of an admin.
func (s *Service) AdminDelete(ctx context.Context, userID uuid.UUID) *apperror.AppError {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("getting user for admin deletion", "error", err, "userID", userID)
		return apperror.Internal(apperror.CodeInternalError, "failed to get user", err)
	}
	if user == nil {
		return apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}
	if user.Status == entity.UserStatusDeleted {
		return apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserDeleted)
	}

	user.SoftDelete()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error("soft-deleting user as admin", "error", err, "userID", userID)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete user", err)
	}

	if err := s.dispatcher.Dispatch(event.UserAccountDeleted{
		UserID:    user.ID,
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching user account deleted event", "error", err)
	}

	return nil
}
