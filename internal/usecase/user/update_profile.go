package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateProfileInput holds the data needed to update a user's profile.
type UpdateProfileInput struct {
	UserID      uuid.UUID
	DisplayName *string
	AvatarURL   *string
	DateOfBirth *string // "2006-01-02" format
	Language    *string
	Country     *string
	Handle      *string
	Pronouns    *string
	Headline    *string
	Bio         *string
	Links       *[]entity.ProfileLink
	Preferences *entity.ProfilePreferences
	Privacy     *entity.ProfilePrivacy
}

// UpdateProfile updates the authenticated user's profile.
func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*entity.User, *apperror.AppError) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		s.log.Error("getting user for profile update", "error", err, "userID", input.UserID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get user", err)
	}
	if user == nil {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserNotFound)
	}
	if user.Status == entity.UserStatusDeleted {
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found", domain.ErrUserDeleted)
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
	if input.Handle != nil {
		if *input.Handle != "" {
			taken, repoErr := s.userRepo.HandleExists(ctx, *input.Handle, input.UserID)
			if repoErr != nil {
				s.log.Error("checking handle existence", "error", repoErr, "userID", input.UserID)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to check handle availability", repoErr)
			}
			if taken {
				return nil, apperror.Conflict(apperror.CodeHandleAlreadyExists, "handle is already taken", nil)
			}
		}
		user.Profile.Handle = input.Handle
	}
	if input.Pronouns != nil {
		user.Profile.Pronouns = input.Pronouns
	}
	if input.Headline != nil {
		user.Profile.Headline = input.Headline
	}
	if input.Bio != nil {
		user.Profile.Bio = input.Bio
	}
	if input.Links != nil {
		user.Profile.Links = *input.Links
	}
	if input.Preferences != nil {
		user.Profile.Preferences = *input.Preferences
	}
	if input.Privacy != nil {
		user.Profile.Privacy = *input.Privacy
	}

	now := time.Now().UTC()
	user.Profile.UpdatedAt = now
	user.UpdatedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error("updating user profile", "error", err, "userID", input.UserID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update profile", err)
	}

	if err := s.dispatcher.Dispatch(event.UserProfileUpdated{
		UserID:    user.ID,
		ActorID:   &user.ID, // self-update
		Timestamp: now,
	}); err != nil {
		s.log.Error("dispatching user profile updated event", "error", err)
	}

	return user, nil
}
