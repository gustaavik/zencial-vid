package repository

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// AuditLogRepository defines persistence operations for audit log entries.
type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	List(ctx context.Context, fs *filter.FilterSet) ([]entity.AuditLog, int64, error)
}
