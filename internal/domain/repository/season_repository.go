package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// SeasonRepository defines persistence operations for series seasons.
type SeasonRepository interface {
	Create(ctx context.Context, season *entity.Season) error
	GetBySeriesAndNumber(ctx context.Context, seriesID uuid.UUID, seasonNumber int) (*entity.Season, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Season, error)
	Update(ctx context.Context, season *entity.Season) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListBySeries(ctx context.Context, seriesID uuid.UUID) ([]entity.Season, error)
}
