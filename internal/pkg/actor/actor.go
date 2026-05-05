// Package actor exposes the authenticated principal's id via context.Context.
//
// Middleware (e.g. JWT auth) writes the actor at the request boundary; use
// cases read it when dispatching domain events for audit purposes. Lives in
// internal/pkg/ so neither domain nor use-case packages have to import
// infrastructure/middleware (Clean Architecture).
package actor

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

var actorKey = contextKey{}

// WithActor returns a copy of ctx carrying the given actor id.
func WithActor(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, actorKey, id)
}

// FromContext returns the actor id stored in ctx, or nil if no actor is set
// (e.g. unauthenticated requests, system-initiated calls like CDN callbacks).
// Returning a pointer makes it easy to assign directly to event ActorID fields.
func FromContext(ctx context.Context) *uuid.UUID {
	id, ok := ctx.Value(actorKey).(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}
