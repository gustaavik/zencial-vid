package service

import (
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// ContentAccessService determines if a user can access specific content.
type ContentAccessService struct{}

// NewContentAccessService creates a new ContentAccessService.
func NewContentAccessService() *ContentAccessService {
	return &ContentAccessService{}
}

// CanAccess checks whether a user can access the given content.
// It verifies subscription status, age restrictions, and content availability.
func (s *ContentAccessService) CanAccess(user *entity.User, content *entity.Content, subscription *entity.Subscription) (bool, string) {
	if !user.IsActive() {
		return false, "user account is not active"
	}

	if !content.IsPlayable() {
		return false, "content is not available for playback"
	}

	if !user.CanAccessContent(content.Rating) {
		return false, "content is restricted by age rating"
	}

	if subscription == nil || !subscription.IsAccessible() {
		return false, "active subscription required"
	}

	return true, ""
}
