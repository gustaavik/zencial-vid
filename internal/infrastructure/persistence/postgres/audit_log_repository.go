package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

var auditLogFilterConfig = filter.Config{
	Columns: map[string]filter.ColumnDef{
		"actor_id":    {DBColumn: "a.actor_id", AllowedOps: []filter.Op{filter.OpEq, filter.OpIn}, Type: filter.TypeUUID},
		"event_name":  {DBColumn: "a.event_name", AllowedOps: []filter.Op{filter.OpEq, filter.OpIn, filter.OpLike}, Type: filter.TypeString},
		"entity_type": {DBColumn: "a.entity_type", AllowedOps: []filter.Op{filter.OpEq, filter.OpIn}, Type: filter.TypeString},
		"entity_id":   {DBColumn: "a.entity_id", AllowedOps: []filter.Op{filter.OpEq}, Type: filter.TypeUUID},
		"occurred_at": {DBColumn: "a.occurred_at", AllowedOps: []filter.Op{filter.OpGte, filter.OpLte}, Type: filter.TypeString},
		// Virtual column joining users table for actor email search.
		"actor_email": {DBColumn: "u.email", AllowedOps: []filter.Op{filter.OpLike, filter.OpEq}, Type: filter.TypeString},
	},
	SortColumns: map[string]filter.SortDef{
		"occurred_at": {DBColumn: "a.occurred_at"},
	},
	DefaultSort: "a.occurred_at DESC",
}

// AuditLogFilterConfig returns the filter configuration for audit logs.
func AuditLogFilterConfig() filter.Config {
	return auditLogFilterConfig
}

// AuditLogRepository implements repository.AuditLogRepository using PostgreSQL.
type AuditLogRepository struct {
	pool *pgxpool.Pool
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

// Create inserts a new audit log row. The audit subscriber must keep this
// path cheap and non-blocking — failures are logged upstream, never propagated
// back to the originating use case.
func (r *AuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	db := connFromCtx(ctx, r.pool)

	metadata := log.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshaling audit metadata: %w", err)
	}

	if log.ID == uuid.Nil {
		log.ID = uuid.Must(uuid.NewV7())
	}

	_, err = db.Exec(ctx, `
		INSERT INTO audit_logs (id, actor_id, event_name, entity_type, entity_id, metadata, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, log.ID, log.ActorID, log.EventName, log.EntityType, log.EntityID, metaJSON, log.OccurredAt)
	if err != nil {
		return fmt.Errorf("creating audit log: %w", err)
	}
	return nil
}

// List returns audit log entries matching the filter set, joined with users
// to populate actor email in a single query.
func (r *AuditLogRepository) List(ctx context.Context, fs *filter.FilterSet) ([]entity.AuditLog, int64, error) {
	db := connFromCtx(ctx, r.pool)

	// Count
	countWhere, countArgs, _ := filter.CountSQL(fs, "", 1)
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM audit_logs a
		LEFT JOIN users u ON a.actor_id = u.id
		%s
	`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting audit logs: %w", err)
	}

	// Data
	sqlRes := filter.ToSQL(fs, "", 1)
	dataQuery := fmt.Sprintf(`
		SELECT a.id, a.actor_id, a.event_name, a.entity_type, a.entity_id, a.metadata, a.occurred_at, a.created_at,
		       COALESCE(u.email, '') AS actor_email
		FROM audit_logs a
		LEFT JOIN users u ON a.actor_id = u.id
		%s %s %s
	`, sqlRes.WhereClause, sqlRes.OrderClause, sqlRes.LimitClause)

	rows, err := db.Query(ctx, dataQuery, sqlRes.Args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit logs: %w", err)
	}
	defer rows.Close()

	var logs []entity.AuditLog
	for rows.Next() {
		var l entity.AuditLog
		var actorID, entityID *uuid.UUID
		var metaBytes []byte

		if err := rows.Scan(
			&l.ID, &actorID, &l.EventName, &l.EntityType, &entityID, &metaBytes, &l.OccurredAt, &l.CreatedAt, &l.ActorEmail,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning audit log: %w", err)
		}
		l.ActorID = actorID
		l.EntityID = entityID
		if len(metaBytes) > 0 {
			if err := json.Unmarshal(metaBytes, &l.Metadata); err != nil {
				return nil, 0, fmt.Errorf("unmarshaling audit metadata: %w", err)
			}
		} else {
			l.Metadata = map[string]any{}
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}
