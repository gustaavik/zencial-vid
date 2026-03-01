package repository

import (
	"context"

	"github.com/google/uuid"
)

// SessionStore defines operations for managing user sessions (refresh tokens).
type SessionStore interface {
	StoreRefreshToken(ctx context.Context, token string, userID uuid.UUID) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (uuid.UUID, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteAllUserTokens(ctx context.Context, userID uuid.UUID) error
}
