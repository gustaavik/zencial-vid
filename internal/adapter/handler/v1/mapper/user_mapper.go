package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// UserToResponse maps a User entity to a UserResponse DTO.
func UserToResponse(user *entity.User) dto.UserResponse {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = string(r)
	}
	resp := dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email.String(),
		Roles:     roles,
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Profile:   ProfileToResponse(&user.Profile),
	}
	return resp
}

// ProfileToResponse maps a UserProfile to a ProfileResponse DTO.
func ProfileToResponse(profile *entity.UserProfile) dto.ProfileResponse {
	links := make([]dto.ProfileLinkDTO, len(profile.Links))
	for i, l := range profile.Links {
		links[i] = dto.ProfileLinkDTO{Label: l.Label, URL: l.URL}
	}

	resp := dto.ProfileResponse{
		DisplayName: profile.DisplayName,
		AvatarURL:   profile.AvatarURL,
		Language:    profile.Language,
		Country:     profile.Country,
		Handle:      profile.Handle,
		Pronouns:    profile.Pronouns,
		Headline:    profile.Headline,
		Bio:         profile.Bio,
		Links:       links,
		Preferences: dto.ProfilePreferencesDTO{
			AllowMatureContent:  profile.Preferences.AllowMatureContent,
			AutoplayNextEpisode: profile.Preferences.AutoplayNextEpisode,
			AlwaysShowSubtitles: profile.Preferences.AlwaysShowSubtitles,
			ShowPaidFirstInFeed: profile.Preferences.ShowPaidFirstInFeed,
		},
		Privacy: dto.ProfilePrivacyDTO{
			ProfileVisibility: profile.Privacy.ProfileVisibility,
			WatchHistory:      profile.Privacy.WatchHistory,
			Watchlist:         profile.Privacy.Watchlist,
			Tipping:           profile.Privacy.Tipping,
		},
	}
	if profile.DateOfBirth != nil {
		dob := profile.DateOfBirth.Format("2006-01-02")
		resp.DateOfBirth = &dob
	}
	return resp
}

// UsersToResponse maps a slice of User entities to UserResponse DTOs.
func UsersToResponse(users []entity.User) []dto.UserResponse {
	result := make([]dto.UserResponse, len(users))
	for i := range users {
		result[i] = UserToResponse(&users[i])
	}
	return result
}
