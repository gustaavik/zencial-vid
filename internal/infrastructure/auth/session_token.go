package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// SessionTokenService generates and hashes opaque session tokens. Tokens are
// 32 bytes of entropy encoded as base64url (no padding, ~43 chars). Hashes are
// SHA-256 hex (64 chars) — deterministic so the bearer token can be looked up
// by hash in O(1) on every request. Bcrypt is unsuitable here because we need
// equality lookup, not password verification.
type SessionTokenService interface {
	// Generate returns a new (raw token, hex SHA-256 hash) pair. The raw token
	// must be returned to the caller exactly once and stored client-side; only
	// the hash is persisted.
	Generate() (token string, hash string, err error)
	// Hash returns the deterministic SHA-256 hex hash of the given token.
	Hash(token string) string
}

type sessionTokenService struct{}

// NewSessionTokenService returns the default SessionTokenService.
func NewSessionTokenService() SessionTokenService {
	return &sessionTokenService{}
}

func (sessionTokenService) Generate() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generating session token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashToken(token), nil
}

func (sessionTokenService) Hash(token string) string {
	return hashToken(token)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
