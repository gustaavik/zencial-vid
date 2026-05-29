package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// SeasonRepository implements repository.SeasonRepository using PostgreSQL.
type SeasonRepository struct {
	pool *pgxpool.Pool
}

// NewSeasonRepository creates a new SeasonRepository.
func NewSeasonRepository(pool *pgxpool.Pool) *SeasonRepository {
	return &SeasonRepository{pool: pool}
}

func (r *SeasonRepository) Create(ctx context.Context, season *entity.Season) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO seasons (
			id, series_id, season_number, season_tag,
			planned_episodes, avg_runtime_secs, release_cadence,
			premiere_date, cadence_day, timezone,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, season.ID, season.SeriesID, season.SeasonNumber, season.SeasonTag,
		season.PlannedEpisodes, season.AvgRuntimeSecs, string(season.ReleaseCadence),
		season.PremiereDate, season.CadenceDay, season.Timezone,
		season.CreatedAt, season.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating season: %w", err)
	}
	return nil
}

func (r *SeasonRepository) GetBySeriesAndNumber(ctx context.Context, seriesID uuid.UUID, seasonNumber int) (*entity.Season, error) {
	db := connFromCtx(ctx, r.pool)
	return scanSeason(db.QueryRow(ctx, `
		SELECT id, series_id, season_number, season_tag,
		       planned_episodes, avg_runtime_secs, release_cadence,
		       premiere_date, cadence_day, timezone, created_at, updated_at
		FROM seasons WHERE series_id = $1 AND season_number = $2
	`, seriesID, seasonNumber))
}

func (r *SeasonRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Season, error) {
	db := connFromCtx(ctx, r.pool)
	return scanSeason(db.QueryRow(ctx, `
		SELECT id, series_id, season_number, season_tag,
		       planned_episodes, avg_runtime_secs, release_cadence,
		       premiere_date, cadence_day, timezone, created_at, updated_at
		FROM seasons WHERE id = $1
	`, id))
}

func (r *SeasonRepository) Update(ctx context.Context, season *entity.Season) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE seasons SET
			season_tag = $2, planned_episodes = $3, avg_runtime_secs = $4,
			release_cadence = $5, premiere_date = $6, cadence_day = $7,
			timezone = $8, updated_at = $9
		WHERE id = $1
	`, season.ID, season.SeasonTag, season.PlannedEpisodes, season.AvgRuntimeSecs,
		string(season.ReleaseCadence), season.PremiereDate, season.CadenceDay,
		season.Timezone, season.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating season: %w", err)
	}
	return nil
}

func (r *SeasonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM seasons WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting season: %w", err)
	}
	return nil
}

func (r *SeasonRepository) ListBySeries(ctx context.Context, seriesID uuid.UUID) ([]entity.Season, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, series_id, season_number, season_tag,
		       planned_episodes, avg_runtime_secs, release_cadence,
		       premiere_date, cadence_day, timezone, created_at, updated_at
		FROM seasons WHERE series_id = $1
		ORDER BY season_number ASC
	`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("listing seasons by series: %w", err)
	}
	defer rows.Close()

	var seasons []entity.Season
	for rows.Next() {
		s, err := scanSeasonRow(rows)
		if err != nil {
			return nil, err
		}
		seasons = append(seasons, *s)
	}
	return seasons, rows.Err()
}

func scanSeason(row pgx.Row) (*entity.Season, error) {
	return scanSeasonRow(row)
}

func scanSeasonRow(row scannable) (*entity.Season, error) {
	var s entity.Season
	var cadence string
	err := row.Scan(
		&s.ID, &s.SeriesID, &s.SeasonNumber, &s.SeasonTag,
		&s.PlannedEpisodes, &s.AvgRuntimeSecs, &cadence,
		&s.PremiereDate, &s.CadenceDay, &s.Timezone,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning season: %w", err)
	}
	s.ReleaseCadence = entity.ReleaseCadence(cadence)
	return &s, nil
}
