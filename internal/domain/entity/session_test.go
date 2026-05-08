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

	// At t+3h, idle slide normally extends to t+4h, but absolute is t+4h so they
	// should be equal. Slide one minute later so idle would push past absolute.
	slid := s.Slide(now.Add(3*time.Hour+10*time.Minute), time.Hour)
	assert.Equal(t, s.AbsoluteExpiresAt, slid.IdleExpiresAt,
		"idle expiry must clamp to absolute")
	// Original is unchanged (immutability).
	assert.Equal(t, now.Add(time.Hour), s.IdleExpiresAt)
}

func TestSession_RevokeReturnsCopy(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewSession(uuid.New(), "h", "", "", "", now, time.Hour, 24*time.Hour)

	r := s.Revoke(now)

	assert.Nil(t, s.RevokedAt, "original must be unchanged")
	assert.NotNil(t, r.RevokedAt)
	assert.Equal(t, now, *r.RevokedAt)
}

func TestSession_ExpiresAtPicksEarlier(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewSession(uuid.New(), "h", "", "", "", now, time.Hour, 24*time.Hour)
	assert.Equal(t, s.IdleExpiresAt, s.ExpiresAt(),
		"idle is the earlier deadline at construction time")

	s.AbsoluteExpiresAt = now.Add(30 * time.Minute) // simulate odd config
	assert.Equal(t, s.AbsoluteExpiresAt, s.ExpiresAt())
}
