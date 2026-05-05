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

// UpdateStatusInput holds the data needed to update a user's status.
type UpdateStatusInput struct {
	UserID uuid.UUID
	Status entity.UserStatus
}

// UpdateStatus updates a user's status (admin operation).
func (s *Service) UpdateStatus(ctx context.Context, input UpdateStatusInput) (*entity.User, *apperror.AppError) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		s.log.Error("getting user for status update", "error", err, "userID", input.UserID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get user", err)
	}
	if user == nil {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}

	if err := s.userRepo.UpdateStatus(ctx, input.UserID, input.Status); err != nil {
		s.log.Error("updating user status", "error", err, "userID", input.UserID, "status", input.Status)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update user status", err)
	}

	user.Status = input.Status
	user.UpdatedAt = time.Now().UTC()

	if err := s.dispatcher.Dispatch(event.UserStatusChanged{
		UserID:    user.ID,
		ActorID:   actor.FromContext(ctx),
		NewStatus: string(input.Status),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching user status changed event", "error", err)
	}

	return user, nil
}
