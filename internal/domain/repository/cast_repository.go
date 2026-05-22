package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// CastRepository defines persistence operations for standalone cast members.
type CastRepository interface {
	Create(ctx context.Context, cast *entity.Cast) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Cast, error)
	GetByName(ctx context.Context, name string) (*entity.Cast, error)
	// FindOrCreate returns the existing cast member with the given name, or
	// creates one if no match exists. Safe for concurrent callers.
	FindOrCreate(ctx context.Context, name string) (*entity.Cast, error)
	Update(ctx context.Context, cast *entity.Cast) error
	Delete(ctx context.Context, id uuid.UUID) error
	// HasVideoWithCaller returns true when the cast member is credited on at
	// least one video uploaded by callerID. Used to authorize publisher edits.
	HasVideoWithCaller(ctx context.Context, castID, callerID uuid.UUID) (bool, error)
	// ListAll returns a paginated list of cast members ordered by name, along
	// with the total count. When includeArchived is false only active members
	// are returned; pass true to include archived members.
	ListAll(ctx context.Context, offset, limit int, includeArchived bool) ([]entity.Cast, int, error)
}
