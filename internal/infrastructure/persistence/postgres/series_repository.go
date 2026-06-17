package postgres

import (
	"context"
	"encoding/json"
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

// seriesColumns is the canonical, ordered column list for series SELECTs.
// The scan order in hydrateSeries must match this exactly.
const seriesColumns = `id, title, slug, description, creator, status, series_type,
	logline, primary_language, origin_country, content_rating, cover_image_key,
	poster_key, banner_key, title_logo_key, uploaded_by, minimum_plan_level,
	autoplay_next, binge_mode, hide_episode_count, default_visibility,
	default_monetization, created_at, updated_at`

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

	monetization, err := marshalMonetization(series.DefaultMonetization)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		INSERT INTO series (id, title, slug, description, creator, status, series_type,
		                    logline, primary_language, origin_country, content_rating,
		                    cover_image_key, poster_key, banner_key, title_logo_key,
		                    uploaded_by, minimum_plan_level, autoplay_next, binge_mode,
		                    hide_episode_count, default_visibility, default_monetization,
		                    created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
		        $17, $18, $19, $20, $21, $22, $23, $24)
	`, series.ID, series.Title, series.Slug.String(), series.Description, series.Creator,
		string(series.Status), string(series.SeriesType), series.Logline,
		series.PrimaryLanguage, series.OriginCountry, series.ContentRating,
		series.CoverImageKey, series.PosterKey, series.BannerKey, series.TitleLogoKey,
		series.UploadedBy, series.MinimumPlanLevel, series.AutoplayNext, series.BingeMode,
		series.HideEpisodeCount, string(series.DefaultVisibility), monetization,
		series.CreatedAt, series.UpdatedAt)
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
	return r.scanSeries(ctx, db, fmt.Sprintf(`SELECT %s FROM series WHERE id = $1`, seriesColumns), id)
}

func (r *SeriesRepository) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Series, error) {
	db := connFromCtx(ctx, r.pool)
	return r.scanSeries(ctx, db, fmt.Sprintf(`SELECT %s FROM series WHERE slug = $1`, seriesColumns), slug.String())
}

func (r *SeriesRepository) Update(ctx context.Context, series *entity.Series) error {
	db := connFromCtx(ctx, r.pool)

	monetization, err := marshalMonetization(series.DefaultMonetization)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		UPDATE series SET title = $2, slug = $3, description = $4, creator = $5,
		       status = $6, series_type = $7, logline = $8, primary_language = $9,
		       origin_country = $10, content_rating = $11, cover_image_key = $12,
		       poster_key = $13, banner_key = $14, title_logo_key = $15,
		       minimum_plan_level = $16, autoplay_next = $17, binge_mode = $18,
		       hide_episode_count = $19, default_visibility = $20,
		       default_monetization = $21, updated_at = $22
		WHERE id = $1
	`, series.ID, series.Title, series.Slug.String(), series.Description, series.Creator,
		string(series.Status), string(series.SeriesType), series.Logline,
		series.PrimaryLanguage, series.OriginCountry, series.ContentRating,
		series.CoverImageKey, series.PosterKey, series.BannerKey, series.TitleLogoKey,
		series.MinimumPlanLevel, series.AutoplayNext, series.BingeMode,
		series.HideEpisodeCount, string(series.DefaultVisibility), monetization,
		series.UpdatedAt)
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
		SELECT %s
		FROM series s
		%s %s %s
	`, seriesColumns, sql.WhereClause, sql.OrderClause, sql.LimitClause)

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
		SELECT %s
		FROM series s
		%s %s %s
	`, seriesColumns, sql.WhereClause, sql.OrderClause, sql.LimitClause)

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
	s, err := hydrateSeries(db.QueryRow(ctx, query, args...).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning series: %w", err)
	}

	genreIDs, err := r.GetGenreIDs(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	s.GenreIDs = genreIDs

	return s, nil
}

func (r *SeriesRepository) scanSeriesRow(rows pgx.Rows) (*entity.Series, error) {
	s, err := hydrateSeries(rows.Scan)
	if err != nil {
		return nil, fmt.Errorf("scanning series row: %w", err)
	}
	return s, nil
}

// hydrateSeries scans a single series row (using the column order in seriesColumns)
// into an entity, decoding enum and JSONB fields. It does not load genre IDs.
func hydrateSeries(scan func(dest ...any) error) (*entity.Series, error) {
	var s entity.Series
	var slug, status, seriesType, visibility string
	var monetization []byte

	if err := scan(
		&s.ID, &s.Title, &slug, &s.Description, &s.Creator, &status,
		&seriesType, &s.Logline, &s.PrimaryLanguage, &s.OriginCountry, &s.ContentRating,
		&s.CoverImageKey, &s.PosterKey, &s.BannerKey, &s.TitleLogoKey,
		&s.UploadedBy, &s.MinimumPlanLevel,
		&s.AutoplayNext, &s.BingeMode, &s.HideEpisodeCount,
		&visibility, &monetization,
		&s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, err
	}

	s.Slug = valueobject.SlugFromTrusted(slug)
	s.Status = entity.SeriesStatus(status)
	s.SeriesType = entity.SeriesType(seriesType)
	s.DefaultVisibility = entity.VideoVisibility(visibility)
	s.DefaultMonetization = []string{}
	if len(monetization) > 0 {
		if err := json.Unmarshal(monetization, &s.DefaultMonetization); err != nil {
			return nil, fmt.Errorf("unmarshaling default_monetization: %w", err)
		}
	}

	return &s, nil
}

// marshalMonetization encodes the default monetization slice as JSONB bytes,
// normalizing nil to an empty array.
func marshalMonetization(monetization []string) ([]byte, error) {
	if monetization == nil {
		monetization = []string{}
	}
	b, err := json.Marshal(monetization)
	if err != nil {
		return nil, fmt.Errorf("marshaling default_monetization: %w", err)
	}
	return b, nil
}
