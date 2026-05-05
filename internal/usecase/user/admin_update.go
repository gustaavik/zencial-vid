package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminUpdateInput holds the fields an admin may update on a user.
// Pointer fields are optional — nil means "no change".
type AdminUpdateInput struct {
	UserID      uuid.UUID
	Email       *string
	Role        *entity.UserRole
	Password    *string
	DisplayName *string
	AvatarURL   *string
	DateOfBirth *string
	Language    *string
	Country     *string
}

// AdminUpdate updates a user's profile, role, email, or password on behalf of an admin.
func (s *Service) AdminUpdate(ctx context.Context, input *AdminUpdateInput) (*entity.User, *apperror.AppError) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		s.log.Error("getting user for admin update", "error", err, "userID", input.UserID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get user", err)
	}
	if user == nil {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}
	if user.Status == entity.UserStatusDeleted {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserDeleted)
	}

	roleChanged := false
	oldRole := user.Role

	if input.Email != nil {
		newEmail, emailErr := valueobject.NewEmail(*input.Email)
		if emailErr != nil {
			return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid email address", emailErr)
		}
		if newEmail.String() != user.Email.String() {
			exists, existsErr := s.userRepo.ExistsByEmail(ctx, newEmail)
			if existsErr != nil {
				s.log.Error("checking email existence", "error", existsErr)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to check email", existsErr)
			}
			if exists {
				return nil, apperror.Conflict(apperror.CodeEmailAlreadyExists, "email already registered", domain.ErrEmailAlreadyExists)
			}
			user.Email = newEmail
		}
	}

	if input.Role != nil && *input.Role != user.Role {
		if *input.Role != entity.RoleAdmin && *input.Role != entity.RoleUser {
			return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid role", nil)
		}
		user.Role = *input.Role
		roleChanged = true
	}

	if input.Password != nil && *input.Password != "" {
		hashed, hashErr := s.hasher.Hash(*input.Password)
		if hashErr != nil {
			s.log.Error("hashing password", "error", hashErr)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to process password", hashErr)
		}
		user.PasswordHash = valueobject.NewHashedPassword(hashed)
	}

	if input.DisplayName != nil {
		user.Profile.DisplayName = *input.DisplayName
	}
	if input.AvatarURL != nil {
		user.Profile.AvatarURL = *input.AvatarURL
	}
	if input.DateOfBirth != nil {
		if *input.DateOfBirth == "" {
			user.Profile.DateOfBirth = nil
		} else {
			dob, parseErr := time.Parse("2006-01-02", *input.DateOfBirth)
			if parseErr != nil {
				return nil, apperror.BadRequest(apperror.CodeInvalidDateFormat, "date of birth must be in YYYY-MM-DD format", parseErr)
			}
			user.Profile.DateOfBirth = &dob
		}
	}
	if input.Language != nil {
		user.Profile.Language = *input.Language
	}
	if input.Country != nil {
		user.Profile.Country = *input.Country
	}

	now := time.Now().UTC()
	user.Profile.UpdatedAt = now
	user.UpdatedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error("updating user as admin", "error", err, "userID", input.UserID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update user", err)
	}

	actorID := actor.FromContext(ctx)

	if err := s.dispatcher.Dispatch(event.UserProfileUpdated{
		UserID:    user.ID,
		ActorID:   actorID,
		Timestamp: now,
	}); err != nil {
		s.log.Error("dispatching user profile updated event", "error", err)
	}

	if roleChanged {
		if err := s.dispatcher.Dispatch(event.UserRoleChanged{
			UserID:    user.ID,
			ActorID:   actorID,
			OldRole:   string(oldRole),
			NewRole:   string(user.Role),
			Timestamp: now,
		}); err != nil {
			s.log.Error("dispatching user role changed event", "error", err)
		}
	}

	return user, nil
}
