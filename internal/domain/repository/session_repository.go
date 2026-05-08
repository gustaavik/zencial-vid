package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// SessionRepository persists user sessions. Implementations return (nil, nil)
// from Get* methods when the row is not found; callers should not rely on a
// sentinel error for that case.
type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, fs *filter.FilterSet) ([]entity.Session, int64, error)
	UpdateActivity(ctx context.Context, id uuid.UUID, lastActivityAt, idleExpiresAt time.Time) error
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error)
	RevokeOthersForUser(ctx context.Context, userID, exceptSessionID uuid.UUID, revokedAt time.Time) (int64, error)
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
