package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

var seriesFilterConfig = filter.Config{
	Columns: map[string]filter.ColumnDef{
		"status":  {DBColumn: "s.status", AllowedOps: []filter.Op{filter.OpEq}, Type: filter.TypeString},
		"title":   {DBColumn: "s.title", AllowedOps: []filter.Op{filter.OpLike}, Type: filter.TypeString},
		"creator": {DBColumn: "s.creator", AllowedOps: []filter.Op{filter.OpLike}, Type: filter.TypeString},
	},
	SortColumns: map[string]filter.SortDef{
		"title":      {DBColumn: "s.title"},
		"created_at": {DBColumn: "s.created_at"},
	},
	DefaultSort: "s.created_at DESC",
}

// SeriesFilterConfig returns the filter configuration for series.
func SeriesFilterConfig() filter.Config {
	return seriesFilterConfig
}

// SeriesRepository implements repository.SeriesRepository using PostgreSQL.
type SeriesRepository struct {
	pool *pgxpool.Pool
}

// NewSeriesRepository creates a new SeriesRepository.
func NewSeriesRepository(pool *pgxpool.Pool) *SeriesRepository {
	return &SeriesRepository{pool: pool}
}

func (r *SeriesRepository) Create(ctx context.Context, series *entity.Series) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		INSERT INTO series (id, title, slug, description, creator, status, cover_image_key,
		                    uploaded_by, minimum_plan_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, series.ID, series.Title, series.Slug.String(), series.Description, series.Creator,
		string(series.Status), series.CoverImageKey, series.UploadedBy,
		series.MinimumPlanLevel, series.CreatedAt, series.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating series: %w", err)
	}

	if len(series.GenreIDs) > 0 {
		if err := r.SetGenres(ctx, series.ID, series.GenreIDs); err != nil {
			return err
		}
	}

	return nil
}

func (r *SeriesRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Series, error) {
	db := connFromCtx(ctx, r.pool)
	return r.scanSeries(ctx, db, `
		SELECT id, title, slug, description, creator, status, cover_image_key,
		       uploaded_by, minimum_plan_level, created_at, updated_at
		FROM series WHERE id = $1
	`, id)
}

func (r *SeriesRepository) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Series, error) {
	db := connFromCtx(ctx, r.pool)
	return r.scanSeries(ctx, db, `
		SELECT id, title, slug, description, creator, status, cover_image_key,
		       uploaded_by, minimum_plan_level, created_at, updated_at
		FROM series WHERE slug = $1
	`, slug.String())
}

func (r *SeriesRepository) Update(ctx context.Context, series *entity.Series) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		UPDATE series SET title = $2, slug = $3, description = $4, creator = $5,
		       status = $6, cover_image_key = $7, minimum_plan_level = $8, updated_at = $9
		WHERE id = $1
	`, series.ID, series.Title, series.Slug.String(), series.Description, series.Creator,
		string(series.Status), series.CoverImageKey, series.MinimumPlanLevel, series.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating series: %w", err)
	}

	return nil
}

func (r *SeriesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM series WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting series: %w", err)
	}
	return nil
}

func (r *SeriesRepository) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, error) {
	return r.listWithBase(ctx, fs, "")
}

func (r *SeriesRepository) ListPublished(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, error) {
	return r.listWithBase(ctx, fs, "s.status = 'published'")
}

func (r *SeriesRepository) ListByUploader(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Series, int64, error) {
	db := connFromCtx(ctx, r.pool)

	baseCondition := "s.uploaded_by = $1"
	baseArgs := []any{uploaderID}

	countWhere, countFilterArgs, _ := filter.CountSQL(fs, baseCondition, 2)
	countArgs := append(append([]any{}, baseArgs...), countFilterArgs...)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM series s %s`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting series by uploader: %w", err)
	}

	sql := filter.ToSQL(fs, baseCondition, 2)
	dataArgs := append(append([]any{}, baseArgs...), sql.Args...)
	dataQuery := fmt.Sprintf(`
		SELECT id, title, slug, description, creator, status, cover_image_key,
		       uploaded_by, minimum_plan_level, created_at, updated_at
		FROM series s
		%s %s %s
	`, sql.WhereClause, sql.OrderClause, sql.LimitClause)

	rows, err := db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing series by uploader: %w", err)
	}
	defer rows.Close()

	return r.collectRows(ctx, rows, total)
}

func (r *SeriesRepository) ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM series WHERE slug = $1)`, slug.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking series slug existence: %w", err)
	}
	return exists, nil
}

func (r *SeriesRepository) SetGenres(ctx context.Context, seriesID uuid.UUID, genreIDs []uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `DELETE FROM series_genres WHERE series_id = $1`, seriesID)
	if err != nil {
		return fmt.Errorf("clearing series genres: %w", err)
	}

	for _, genreID := range genreIDs {
		_, err := db.Exec(ctx, `
			INSERT INTO series_genres (series_id, genre_id) VALUES ($1, $2)
		`, seriesID, genreID)
		if err != nil {
			return fmt.Errorf("setting series genre: %w", err)
		}
	}

	return nil
}

func (r *SeriesRepository) GetGenreIDs(ctx context.Context, seriesID uuid.UUID) ([]uuid.UUID, error) {
	db := connFromCtx(ctx, r.pool)

	rows, err := db.Query(ctx, `SELECT genre_id FROM series_genres WHERE series_id = $1`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("getting series genre IDs: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning genre ID: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *SeriesRepository) listWithBase(ctx context.Context, fs *filter.FilterSet, baseCondition string) ([]entity.Series, int64, error) {
	db := connFromCtx(ctx, r.pool)

	countWhere, countArgs, _ := filter.CountSQL(fs, baseCondition, 1)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM series s %s`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting series: %w", err)
	}

	sql := filter.ToSQL(fs, baseCondition, 1)
	dataQuery := fmt.Sprintf(`
		SELECT id, title, slug, description, creator, status, cover_image_key,
		       uploaded_by, minimum_plan_level, created_at, updated_at
		FROM series s
		%s %s %s
	`, sql.WhereClause, sql.OrderClause, sql.LimitClause)

	rows, err := db.Query(ctx, dataQuery, sql.Args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing series: %w", err)
	}
	defer rows.Close()

	return r.collectRows(ctx, rows, total)
}

func (r *SeriesRepository) collectRows(ctx context.Context, rows pgx.Rows, total int64) ([]entity.Series, int64, error) {
	var series []entity.Series
	for rows.Next() {
		s, err := r.scanSeriesRow(rows)
		if err != nil {
			return nil, 0, err
		}
		series = append(series, *s)
	}

	for i := range series {
		genreIDs, err := r.GetGenreIDs(ctx, series[i].ID)
		if err != nil {
			return nil, 0, err
		}
		series[i].GenreIDs = genreIDs
	}

	return series, total, nil
}

func (r *SeriesRepository) scanSeries(ctx context.Context, db DBTX, query string, args ...any) (*entity.Series, error) {
	var s entity.Series
	var slug, status string

	err := db.QueryRow(ctx, query, args...).Scan(
		&s.ID, &s.Title, &slug, &s.Description, &s.Creator,
		&status, &s.CoverImageKey, &s.UploadedBy,
		&s.MinimumPlanLevel, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning series: %w", err)
	}

	s.Slug = valueobject.SlugFromTrusted(slug)
	s.Status = entity.SeriesStatus(status)

	genreIDs, err := r.GetGenreIDs(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	s.GenreIDs = genreIDs

	return &s, nil
}

func (r *SeriesRepository) scanSeriesRow(rows pgx.Rows) (*entity.Series, error) {
	var s entity.Series
	var slug, status string

	err := rows.Scan(
		&s.ID, &s.Title, &slug, &s.Description, &s.Creator,
		&status, &s.CoverImageKey, &s.UploadedBy,
		&s.MinimumPlanLevel, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning series row: %w", err)
	}

	s.Slug = valueobject.SlugFromTrusted(slug)
	s.Status = entity.SeriesStatus(status)

	return &s, nil
}
