package audit

import (
	"context"
	"log/slog"
	"time"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// persistTimeout caps how long an audit insert can take before we bail out
// and log the failure. Audit must never block the originating use case.
const persistTimeout = 5 * time.Second

// Register subscribes a single audit handler that persists every event
// implementing event.AuditableEvent. Events that don't implement the
// interface are skipped with a debug log.
func Register(d event.Dispatcher, repo repository.AuditLogRepository, log *slog.Logger) {
	d.SubscribeAll(func(e event.Event) error {
		auditable, ok := e.(event.AuditableEvent)
		if !ok {
			log.Debug("event is not auditable, skipping",
				"event", e.EventName(),
			)
			return nil
		}

		entry := &entity.AuditLog{
			ActorID:    auditable.AuditActor(),
			EventName:  auditable.EventName(),
			EntityType: auditable.AuditEntityType(),
			EntityID:   auditable.AuditEntityID(),
			Metadata:   auditable.AuditMetadata(),
			OccurredAt: auditable.OccurredAt(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), persistTimeout)
		defer cancel()

		if err := repo.Create(ctx, entry); err != nil {
			log.Error("failed to persist audit log",
				"event", e.EventName(),
				"entity_type", entry.EntityType,
				"error", err,
			)
			// Swallow: audit failures must not break the originating operation.
			return nil
		}
		return nil
	})
}
