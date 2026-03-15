package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// PlanRepository defines persistence operations for plans.
type PlanRepository interface {
	Create(ctx context.Context, plan *entity.Plan) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error)
	GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Plan, error)
	Update(ctx context.Context, plan *entity.Plan) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, fs filter.FilterSet) ([]entity.Plan, int64, error)
	ListActive(ctx context.Context) ([]entity.Plan, error)
	ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error)
}
