package user

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Service handles user profile use cases.
type Service struct {
	userRepo repository.UserRepository
	log      *slog.Logger
}

// NewService creates a new user Service.
func NewService(userRepo repository.UserRepository, log *slog.Logger) *Service {
	return &Service{
		userRepo: userRepo,
		log:      log,
	}
}

// GetProfile retrieves a user's profile by ID.
func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.User, *apperror.AppError) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("getting user profile", "error", err, "user_id", userID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get profile", err)
	}
	if user == nil {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}
	return user, nil
}

// UpdateProfileInput holds the fields to update on a user profile.
type UpdateProfileInput struct {
	DisplayName *string
	AvatarURL   *string
	DateOfBirth *time.Time
	Language    *string
	Country     *string
}

// UpdateProfile updates a user's profile.
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*entity.User, *apperror.AppError) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("getting user for update", "error", err, "user_id", userID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get user", err)
	}
	if user == nil {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}

	if input.DisplayName != nil {
		user.Profile.DisplayName = *input.DisplayName
	}
	if input.AvatarURL != nil {
		user.Profile.AvatarURL = *input.AvatarURL
	}
	if input.DateOfBirth != nil {
		user.Profile.DateOfBirth = input.DateOfBirth
	}
	if input.Language != nil {
		user.Profile.Language = *input.Language
	}
	if input.Country != nil {
		user.Profile.Country = *input.Country
	}

	now := time.Now()
	user.Profile.UpdatedAt = now
	user.UpdatedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error("updating user profile", "error", err, "user_id", userID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update profile", err)
	}

	return user, nil
}

// DeleteAccount soft-deletes a user account.
func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID) *apperror.AppError {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.log.Error("deleting user account", "error", err, "user_id", userID)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete account", err)
	}
	return nil
}

// ListUsers lists all users (admin use case).
func (s *Service) ListUsers(ctx context.Context, page, perPage int) ([]entity.User, int64, *apperror.AppError) {
	users, total, err := s.userRepo.List(ctx, page, perPage)
	if err != nil {
		s.log.Error("listing users", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list users", err)
	}
	return users, total, nil
}

// UpdateUserStatus updates a user's status (admin use case).
func (s *Service) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status entity.UserStatus) *apperror.AppError {
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		s.log.Error("updating user status", "error", err, "user_id", userID)
		return apperror.Internal(apperror.CodeInternalError, "failed to update user status", err)
	}
	return nil
}
