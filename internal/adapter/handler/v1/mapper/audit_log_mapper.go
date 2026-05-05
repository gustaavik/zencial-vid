package mapper

import (
	"time"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// AuditLogToResponse converts a domain AuditLog into its API representation.
func AuditLogToResponse(l *entity.AuditLog) dto.AuditLogResponse {
	resp := dto.AuditLogResponse{
		ID:         l.ID.String(),
		EventName:  l.EventName,
		EntityType: l.EntityType,
		Metadata:   l.Metadata,
		OccurredAt: l.OccurredAt.UTC().Format(time.RFC3339),
	}
	if l.ActorID != nil {
		resp.Actor = &dto.AuditActorRef{
			ID:    l.ActorID.String(),
			Email: l.ActorEmail,
		}
	}
	if l.EntityID != nil {
		s := l.EntityID.String()
		resp.EntityID = &s
	}
	return resp
}

// AuditLogsToResponse converts a slice of domain AuditLogs.
func AuditLogsToResponse(logs []entity.AuditLog) []dto.AuditLogResponse {
	out := make([]dto.AuditLogResponse, len(logs))
	for i := range logs {
		out[i] = AuditLogToResponse(&logs[i])
	}
	return out
}
