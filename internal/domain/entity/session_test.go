package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewSession_TruncatesUserAgent(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	long := strings.Repeat("x", userAgentMaxLen+50)

	s := NewSession(uuid.New(), "hash", "iPhone", long, "1.2.3.4", now, time.Hour, 24*time.Hour)

	assert.Equal(t, userAgentMaxLen, len(s.UserAgent))
	assert.Equal(t, now, s.CreatedAt)
	assert.Equal(t, now, s.LastActivityAt)
	assert.Equal(t, now.Add(time.Hour), s.IdleExpiresAt)
	assert.Equal(t, now.Add(24*time.Hour), s.AbsoluteExpiresAt)
	assert.Nil(t, s.RevokedAt)
}

func TestSession_IsExpiredAndIsActive(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewSession(uuid.New(), "h", "", "", "", now, time.Hour, 24*time.Hour)

	cases := []struct {
		name     string
		at       time.Time
		expired  bool
		isActive bool
		revoke   bool
	}{
		{"fresh", now, false, true, false},
		{"middle", now.Add(30 * time.Minute), false, true, false},
		{"past idle", now.Add(2 * time.Hour), true, false, false},
		{"past absolute", now.Add(48 * time.Hour), true, false, false},
		{"revoked-fresh", now, false, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess := *s
			if tc.revoke {
				rev := sess.Revoke(now)
				sess = *rev
			}
			assert.Equal(t, tc.expired, sess.IsExpired(tc.at))
			assert.Equal(t, tc.isActive, sess.IsActive(tc.at))
		})
	}
}

func TestSession_SlideClampsToAbsoluteExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewSession(uuid.New(), "h", "", "", "", now, time.Hour, 4*time.Hour)
	absoluteAt := s.AbsoluteExpiresAt
	slideAt := now.Add(3*time.Hour + 10*time.Minute)

	// Sliding past the absolute deadline must clamp idle to absolute, not push
	// it past. Slide mutates in place and returns the same pointer.
	slid := s.Slide(slideAt, time.Hour)
	assert.Same(t, s, slid, "Slide should return the same pointer (in-place)")
	assert.Equal(t, absoluteAt, slid.IdleExpiresAt,
		"idle expiry must clamp to absolute")
	assert.Equal(t, slideAt, s.LastActivityAt)
}

func TestSession_RevokeMarksSession(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewSession(uuid.New(), "h", "", "", "", now, time.Hour, 24*time.Hour)

	r := s.Revoke(now)

	assert.Same(t, s, r, "Revoke should return the same pointer (in-place)")
	assert.NotNil(t, s.RevokedAt)
	assert.Equal(t, now, *s.RevokedAt)
	assert.True(t, s.IsRevoked())
}

func TestSession_ExpiresAtPicksEarlier(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewSession(uuid.New(), "h", "", "", "", now, time.Hour, 24*time.Hour)
	assert.Equal(t, s.IdleExpiresAt, s.ExpiresAt(),
		"idle is the earlier deadline at construction time")

	s.AbsoluteExpiresAt = now.Add(30 * time.Minute) // simulate odd config
	assert.Equal(t, s.AbsoluteExpiresAt, s.ExpiresAt())
}
