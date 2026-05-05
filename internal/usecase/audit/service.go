// Package audit provides query and persistence services for the admin audit
// log. The subscriber registered via Register persists every dispatched
// AuditableEvent; the Service exposes a read API for admin handlers.
package audit

import (
	"context"
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// Service handles audit log read operations.
type Service struct {
	repo repository.AuditLogRepository
	log  *slog.Logger
}

// NewService creates a new audit Service.
func NewService(repo repository.AuditLogRepository, log *slog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// List returns audit log entries matching the filter set.
func (s *Service) List(ctx context.Context, fs *filter.FilterSet) ([]entity.AuditLog, int64, *apperror.AppError) {
	logs, total, err := s.repo.List(ctx, fs)
	if err != nil {
		s.log.Error("listing audit logs", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list audit logs", err)
	}
	return logs, total, nil
}
