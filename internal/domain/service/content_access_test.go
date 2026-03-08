package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func activeUser() *entity.User {
	email := valueobject.EmailFromTrusted("user@example.com")
	u := entity.NewUser(email, valueobject.NewHashedPassword("hash"))
	return u
}

func playableContent() *entity.BaseContent {
	slug := valueobject.SlugFromTrusted("test-film")
	return &entity.BaseContent{
		ID:     uuid.New(),
		Type:   entity.ContentTypeFilm,
		Title:  "Test Film",
		Slug:   slug,
		Rating: valueobject.RatingPG,
		Status: entity.ContentStatusPublished,
		Asset: &entity.VideoAsset{
			ID:     uuid.New(),
			Status: entity.VideoAssetReady,
		},
	}
}

// playableContentWithPlan returns playable content that requires a subscription plan.
func playableContentWithPlan() *entity.BaseContent {
	c := playableContent()
	c.Plan = &entity.Plan{ID: uuid.New(), Name: "Basic", Tier: entity.PlanBasic}
	return c
}

func activeSubscription() *entity.Subscription {
	return &entity.Subscription{
		ID:                 uuid.New(),
		UserID:             uuid.New(),
		PlanID:             uuid.New(),
		Status:             entity.SubscriptionActive,
		CurrentPeriodStart: time.Now().Add(-7 * 24 * time.Hour),
		CurrentPeriodEnd:   time.Now().Add(23 * 24 * time.Hour),
	}
}

func TestContentAccessService_CanAccess(t *testing.T) {
	svc := NewContentAccessService()

	t.Run("active user with playable content and active subscription can access", func(t *testing.T) {
		ok, reason := svc.CanAccess(activeUser(), playableContent(), activeSubscription())
		assert.True(t, ok)
		assert.Empty(t, reason)
	})

	t.Run("suspended user cannot access", func(t *testing.T) {
		user := activeUser()
		user.Suspend()
		ok, reason := svc.CanAccess(user, playableContent(), activeSubscription())
		assert.False(t, ok)
		assert.Contains(t, reason, "not active")
	})

	t.Run("deleted user cannot access", func(t *testing.T) {
		user := activeUser()
		user.SoftDelete()
		ok, reason := svc.CanAccess(user, playableContent(), activeSubscription())
		assert.False(t, ok)
		assert.Contains(t, reason, "not active")
	})

	t.Run("unpublished content cannot be accessed", func(t *testing.T) {
		content := playableContent()
		content.Status = entity.ContentStatusDraft
		ok, reason := svc.CanAccess(activeUser(), content, activeSubscription())
		assert.False(t, ok)
		assert.Contains(t, reason, "not available")
	})

	t.Run("age restricted content denied for user without DOB", func(t *testing.T) {
		content := playableContent()
		content.Rating = valueobject.RatingR // 17+
		user := activeUser()
		// user has no DOB, so only unrestricted content is allowed
		ok, reason := svc.CanAccess(user, content, activeSubscription())
		assert.False(t, ok)
		assert.Contains(t, reason, "age rating")
	})

	t.Run("age restricted content allowed for adult user", func(t *testing.T) {
		content := playableContent()
		content.Rating = valueobject.RatingR
		user := activeUser()
		dob := time.Now().AddDate(-25, 0, 0) // 25 years old
		user.Profile.DateOfBirth = &dob
		ok, reason := svc.CanAccess(user, content, activeSubscription())
		assert.True(t, ok)
		assert.Empty(t, reason)
	})

	t.Run("no subscription denies access to paid content", func(t *testing.T) {
		ok, reason := svc.CanAccess(activeUser(), playableContentWithPlan(), nil)
		assert.False(t, ok)
		assert.Contains(t, reason, "subscription required")
	})

	t.Run("expired subscription denies access to paid content", func(t *testing.T) {
		sub := activeSubscription()
		sub.Status = entity.SubscriptionExpired
		ok, reason := svc.CanAccess(activeUser(), playableContentWithPlan(), sub)
		assert.False(t, ok)
		assert.Contains(t, reason, "subscription required")
	})

	t.Run("free content accessible without subscription", func(t *testing.T) {
		content := playableContent()
		content.Plan = nil // explicitly free
		ok, reason := svc.CanAccess(activeUser(), content, nil)
		assert.True(t, ok)
		assert.Empty(t, reason)
	})
}
