package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

var sessionFilterConfig = filter.Config{
	Columns: map[string]filter.ColumnDef{
		"user_id": {DBColumn: "s.user_id", AllowedOps: []filter.Op{filter.OpEq}, Type: filter.TypeUUID},
	},
	SortColumns: map[string]filter.SortDef{
		"created_at":       {DBColumn: "s.created_at"},
		"last_activity_at": {DBColumn: "s.last_activity_at"},
	},
	DefaultSort: "s.last_activity_at DESC",
}

// SessionFilterConfig returns the filter configuration for sessions.
func SessionFilterConfig() filter.Config {
	return sessionFilterConfig
}

// SessionRepository implements repository.SessionRepository using PostgreSQL.
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

const sessionSelectColumns = `
	s.id, s.user_id, s.token_hash, s.device_name, s.user_agent, s.ip_address,
	s.created_at, s.last_activity_at, s.idle_expires_at, s.absolute_expires_at, s.revoked_at`

func (r *SessionRepository) Create(ctx context.Context, s *entity.Session) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO user_sessions (
			id, user_id, token_hash, device_name, user_agent, ip_address,
			created_at, last_activity_at, idle_expires_at, absolute_expires_at, revoked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, s.ID, s.UserID, s.TokenHash, s.DeviceName, s.UserAgent, s.IPAddress,
		s.CreatedAt, s.LastActivityAt, s.IdleExpiresAt, s.AbsoluteExpiresAt, nullTime(s.RevokedAt))
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error) {
	db := connFromCtx(ctx, r.pool)
	row := db.QueryRow(ctx, `SELECT `+sessionSelectColumns+` FROM user_sessions s WHERE s.token_hash = $1`, tokenHash)
	return scanSession(row)
}

func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	db := connFromCtx(ctx, r.pool)
	row := db.QueryRow(ctx, `SELECT `+sessionSelectColumns+` FROM user_sessions s WHERE s.id = $1`, id)
	return scanSession(row)
}

func (r *SessionRepository) ListByUserID(
	ctx context.Context,
	userID uuid.UUID,
	fs *filter.FilterSet,
) ([]entity.Session, int64, error) {
	db := connFromCtx(ctx, r.pool)
	baseCondition := "s.user_id = $1 AND s.revoked_at IS NULL"

	countWhere, countArgs, _ := filter.CountSQL(fs, baseCondition, 2)
	countArgs = append([]any{userID}, countArgs...)
	var total int64
	if err := db.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM user_sessions s %s`, countWhere),
		countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting sessions: %w", err)
	}

	sqlFilter := filter.ToSQL(fs, baseCondition, 2)
	args := append([]any{userID}, sqlFilter.Args...)
	query := fmt.Sprintf(`
		SELECT %s
		FROM user_sessions s
		%s %s %s
	`, sessionSelectColumns, sqlFilter.WhereClause, sqlFilter.OrderClause, sqlFilter.LimitClause)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing sessions: %w", err)
	}
	defer rows.Close()

	var sessions []entity.Session
	for rows.Next() {
		s, err := scanSessionRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning session: %w", err)
		}
		sessions = append(sessions, *s)
	}
	return sessions, total, nil
}

func (r *SessionRepository) UpdateActivity(
	ctx context.Context,
	id uuid.UUID,
	lastActivityAt, idleExpiresAt time.Time,
) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE user_sessions
		SET last_activity_at = $2, idle_expires_at = $3
		WHERE id = $1 AND revoked_at IS NULL
	`, id, lastActivityAt, idleExpiresAt)
	if err != nil {
		return fmt.Errorf("updating session activity: %w", err)
	}
	return nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE user_sessions SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL
	`, id, revokedAt)
	if err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

func (r *SessionRepository) RevokeAllForUser(
	ctx context.Context,
	userID uuid.UUID,
	revokedAt time.Time,
) (int64, error) {
	db := connFromCtx(ctx, r.pool)
	tag, err := db.Exec(ctx, `
		UPDATE user_sessions SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, revokedAt)
	if err != nil {
		return 0, fmt.Errorf("revoking all sessions for user: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *SessionRepository) RevokeOthersForUser(
	ctx context.Context,
	userID, exceptSessionID uuid.UUID,
	revokedAt time.Time,
) (int64, error) {
	db := connFromCtx(ctx, r.pool)
	tag, err := db.Exec(ctx, `
		UPDATE user_sessions SET revoked_at = $3
		WHERE user_id = $1 AND id <> $2 AND revoked_at IS NULL
	`, userID, exceptSessionID, revokedAt)
	if err != nil {
		return 0, fmt.Errorf("revoking other sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	db := connFromCtx(ctx, r.pool)
	tag, err := db.Exec(ctx, `
		DELETE FROM user_sessions
		WHERE absolute_expires_at < $1
		   OR (revoked_at IS NOT NULL AND revoked_at < $1)
	`, before)
	if err != nil {
		return 0, fmt.Errorf("deleting expired sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}

func scanSession(row pgx.Row) (*entity.Session, error) {
	s, err := scanSessionRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting session: %w", err)
	}
	return s, nil
}

// rowScanner abstracts pgx.Row and pgx.Rows so scanSessionRow can serve both
// QueryRow and Query call sites.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanSessionRow(row rowScanner) (*entity.Session, error) {
	var s entity.Session
	var revokedAt sql.NullTime
	if err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.TokenHash,
		&s.DeviceName,
		&s.UserAgent,
		&s.IPAddress,
		&s.CreatedAt,
		&s.LastActivityAt,
		&s.IdleExpiresAt,
		&s.AbsoluteExpiresAt,
		&revokedAt,
	); err != nil {
		return nil, err
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		s.RevokedAt = &t
	}
	return &s, nil
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}
