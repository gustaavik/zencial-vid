package entity

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session backed by an opaque token.
// The raw token is never persisted — only the SHA-256 hash is stored, and
// equality lookup at request time is done by hashing the bearer token and
// matching against TokenHash.
type Session struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	TokenHash         string
	DeviceName        string
	UserAgent         string
	IPAddress         string
	CreatedAt         time.Time
	LastActivityAt    time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	RevokedAt         *time.Time
}

// userAgentMaxLen mirrors the user_agent column width.
const userAgentMaxLen = 512

// NewSession constructs a session at creation time. UserAgent is truncated to
// the column width so callers don't have to worry about oversize headers.
func NewSession(
	userID uuid.UUID,
	tokenHash, deviceName, userAgent, ipAddr string,
	now time.Time,
	idleTimeout, absoluteTimeout time.Duration,
) *Session {
	return &Session{
		ID:                uuid.Must(uuid.NewV7()),
		UserID:            userID,
		TokenHash:         tokenHash,
		DeviceName:        deviceName,
		UserAgent:         truncate(userAgent, userAgentMaxLen),
		IPAddress:         ipAddr,
		CreatedAt:         now,
		LastActivityAt:    now,
		IdleExpiresAt:     now.Add(idleTimeout),
		AbsoluteExpiresAt: now.Add(absoluteTimeout),
	}
}

// IsRevoked reports whether the session has been explicitly revoked.
func (s *Session) IsRevoked() bool { return s.RevokedAt != nil }

// IsExpired reports whether the session is past its idle or absolute deadline.
func (s *Session) IsExpired(now time.Time) bool {
	return !now.Before(s.IdleExpiresAt) || !now.Before(s.AbsoluteExpiresAt)
}

// IsActive reports whether the session can authorize a request.
func (s *Session) IsActive(now time.Time) bool {
	return !s.IsRevoked() && !s.IsExpired(now)
}

// Slide returns a copy with LastActivityAt and IdleExpiresAt advanced.
// AbsoluteExpiresAt is never moved; if sliding would push the idle deadline
// past absolute, idle is clamped to absolute so the hard cap is preserved.
func (s *Session) Slide(now time.Time, idleTimeout time.Duration) *Session {
	newIdle := now.Add(idleTimeout)
	if newIdle.After(s.AbsoluteExpiresAt) {
		newIdle = s.AbsoluteExpiresAt
	}
	s.LastActivityAt = now
	s.IdleExpiresAt = newIdle
	return s
}

// Revoke returns a copy marked revoked at the given time.
func (s *Session) Revoke(now time.Time) *Session {
	s.RevokedAt = &now
	return s
}

// ExpiresAt returns the effective deadline (the earlier of idle and absolute).
func (s *Session) ExpiresAt() time.Time {
	if s.AbsoluteExpiresAt.Before(s.IdleExpiresAt) {
		return s.AbsoluteExpiresAt
	}
	return s.IdleExpiresAt
}

func truncate(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}
