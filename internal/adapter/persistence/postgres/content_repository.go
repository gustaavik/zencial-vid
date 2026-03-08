package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// ContentRepository implements repository.ContentRepository using PostgreSQL.
type ContentRepository struct {
	pool *pgxpool.Pool
}

// NewContentRepository creates a new ContentRepository.
func NewContentRepository(pool *pgxpool.Pool) *ContentRepository {
	return &ContentRepository{pool: pool}
}

// ─── Film ────────────────────────────────────────────────────────────────────

func (r *ContentRepository) CreateFilm(ctx context.Context, film *entity.Film) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO content (id, type, title, slug, description, synopsis, rating, release_year,
		                     poster_url, backdrop_url, trailer_url, director, status, is_featured,
		                     genre_id, plan_id, created_at, updated_at)
		VALUES ($1, 'film', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`, film.ID, film.Title, film.Slug.String(), film.Description, film.Synopsis,
		film.Rating, film.ReleaseYear, film.PosterURL, film.BackdropURL,
		film.TrailerURL, film.Director, film.Status, film.IsFeatured,
		genreID(film.Genre), planID(film.Plan),
		film.CreatedAt, film.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "content_slug_key" {
			return domain.ErrSlugAlreadyExists
		}
		return fmt.Errorf("creating film content row: %w", err)
	}
	_, err = db.Exec(ctx, `
		INSERT INTO films (content_id, duration_seconds) VALUES ($1, $2)
		ON CONFLICT (content_id) DO UPDATE SET duration_seconds = EXCLUDED.duration_seconds
	`, film.ID, film.Duration.Seconds)
	if err != nil {
		return fmt.Errorf("creating film row: %w", err)
	}
	return nil
}

func (r *ContentRepository) GetFilmByID(ctx context.Context, id uuid.UUID) (*entity.Film, error) {
	return r.loadFilm(ctx, `WHERE c.id = $1`, id)
}

func (r *ContentRepository) GetFilmBySlug(ctx context.Context, slug string) (*entity.Film, error) {
	return r.loadFilm(ctx, `WHERE c.slug = $1 AND c.status = 'published'`, slug)
}

func (r *ContentRepository) UpdateFilm(ctx context.Context, film *entity.Film) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE content SET title=$2, slug=$3, description=$4, synopsis=$5, rating=$6,
		       release_year=$7, poster_url=$8, backdrop_url=$9, trailer_url=$10,
		       director=$11, status=$12, is_featured=$13, genre_id=$14, plan_id=$15, updated_at=$16
		WHERE id = $1
	`, film.ID, film.Title, film.Slug.String(), film.Description, film.Synopsis,
		film.Rating, film.ReleaseYear, film.PosterURL, film.BackdropURL,
		film.TrailerURL, film.Director, film.Status, film.IsFeatured,
		genreID(film.Genre), planID(film.Plan), film.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating film content row: %w", err)
	}
	_, err = db.Exec(ctx, `
		UPDATE films SET duration_seconds=$2 WHERE content_id = $1
	`, film.ID, film.Duration.Seconds)
	if err != nil {
		return fmt.Errorf("updating film row: %w", err)
	}
	return nil
}

func (r *ContentRepository) loadFilm(ctx context.Context, where string, args ...any) (*entity.Film, error) {
	db := connFromCtx(ctx, r.pool)
	var (
		f          entity.Film
		slug       string
		durSec     int64
		genreIDVal *uuid.UUID
		genreName  *string
		genreSlug  *string
		planIDVal  *uuid.UUID
		planName   *string
		planTier   *string
	)
	err := db.QueryRow(ctx, `
		SELECT c.id, c.title, c.slug, c.description, c.synopsis, c.rating,
		       c.release_year, c.poster_url, c.backdrop_url, c.trailer_url, c.director,
		       c.status, c.is_featured, c.created_at, c.updated_at,
		       COALESCE(f.duration_seconds, 0),
		       c.genre_id, g.name, g.slug,
		       c.plan_id, p.name, p.tier
		FROM content c
		LEFT JOIN films  f ON f.content_id = c.id
		LEFT JOIN genres g ON g.id = c.genre_id
		LEFT JOIN plans  p ON p.id = c.plan_id
		`+where, args...).Scan(
		&f.ID, &f.Title, &slug, &f.Description, &f.Synopsis, &f.Rating,
		&f.ReleaseYear, &f.PosterURL, &f.BackdropURL, &f.TrailerURL, &f.Director,
		&f.Status, &f.IsFeatured, &f.CreatedAt, &f.UpdatedAt,
		&durSec,
		&genreIDVal, &genreName, &genreSlug,
		&planIDVal, &planName, &planTier,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("loading film: %w", err)
	}
	f.Slug = valueobject.SlugFromTrusted(slug)
	f.Duration = valueobject.NewDuration(durSec)
	f.Type = entity.ContentTypeFilm
	if genreIDVal != nil {
		f.Genre = &entity.Genre{ID: *genreIDVal, Name: *genreName, Slug: *genreSlug}
	}
	if planIDVal != nil {
		f.Plan = &entity.Plan{ID: *planIDVal, Name: *planName, Tier: entity.PlanTier(*planTier)}
	}
	return &f, nil
}

// ─── Video ───────────────────────────────────────────────────────────────────

func (r *ContentRepository) CreateVideo(ctx context.Context, video *entity.Video) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO content (id, type, title, slug, description, synopsis, rating,
		                     poster_url, status, is_featured, genre_id, plan_id, created_at, updated_at)
		VALUES ($1, 'video', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, video.ID, video.Title, video.Slug.String(), video.Description, video.Synopsis,
		video.Rating, video.PosterURL, video.Status, video.IsFeatured,
		genreID(video.Genre), planID(video.Plan),
		video.CreatedAt, video.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "content_slug_key" {
			return domain.ErrSlugAlreadyExists
		}
		return fmt.Errorf("creating video content row: %w", err)
	}
	uploadedAt := video.UploadedAt
	if uploadedAt.IsZero() {
		uploadedAt = time.Now()
	}
	_, err = db.Exec(ctx, `
		INSERT INTO videos (content_id, duration_seconds, creator_name, uploaded_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (content_id) DO UPDATE
		  SET duration_seconds = EXCLUDED.duration_seconds,
		      creator_name     = EXCLUDED.creator_name,
		      uploaded_at      = EXCLUDED.uploaded_at
	`, video.ID, video.Duration.Seconds, video.CreatorName, uploadedAt)
	if err != nil {
		return fmt.Errorf("creating video row: %w", err)
	}
	return nil
}

func (r *ContentRepository) GetVideoByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	return r.loadVideo(ctx, `WHERE c.id = $1`, id)
}

func (r *ContentRepository) GetVideoBySlug(ctx context.Context, slug string) (*entity.Video, error) {
	return r.loadVideo(ctx, `WHERE c.slug = $1 AND c.status = 'published'`, slug)
}

func (r *ContentRepository) UpdateVideo(ctx context.Context, video *entity.Video) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE content SET title=$2, slug=$3, description=$4, synopsis=$5, rating=$6,
		       poster_url=$7, status=$8, is_featured=$9, genre_id=$10, plan_id=$11, updated_at=$12
		WHERE id = $1
	`, video.ID, video.Title, video.Slug.String(), video.Description, video.Synopsis,
		video.Rating, video.PosterURL, video.Status, video.IsFeatured,
		genreID(video.Genre), planID(video.Plan), video.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating video content row: %w", err)
	}
	_, err = db.Exec(ctx, `
		UPDATE videos SET duration_seconds=$2, creator_name=$3 WHERE content_id = $1
	`, video.ID, video.Duration.Seconds, video.CreatorName)
	if err != nil {
		return fmt.Errorf("updating video row: %w", err)
	}
	return nil
}

func (r *ContentRepository) loadVideo(ctx context.Context, where string, args ...any) (*entity.Video, error) {
	db := connFromCtx(ctx, r.pool)
	var (
		v           entity.Video
		slug        string
		durSec      int64
		uploadedAt  time.Time
		genreIDVal  *uuid.UUID
		genreName   *string
		genreSlug   *string
		planIDVal   *uuid.UUID
		planName    *string
		planTier    *string
	)
	err := db.QueryRow(ctx, `
		SELECT c.id, c.title, c.slug, c.description, c.synopsis, c.rating,
		       c.poster_url, c.status, c.is_featured, c.created_at, c.updated_at,
		       COALESCE(v.duration_seconds, 0), COALESCE(v.creator_name, ''),
		       COALESCE(v.uploaded_at, c.created_at),
		       c.genre_id, g.name, g.slug,
		       c.plan_id, p.name, p.tier
		FROM content c
		LEFT JOIN videos v ON v.content_id = c.id
		LEFT JOIN genres g ON g.id = c.genre_id
		LEFT JOIN plans  p ON p.id = c.plan_id
		`+where, args...).Scan(
		&v.ID, &v.Title, &slug, &v.Description, &v.Synopsis, &v.Rating,
		&v.PosterURL, &v.Status, &v.IsFeatured, &v.CreatedAt, &v.UpdatedAt,
		&durSec, &v.CreatorName, &uploadedAt,
		&genreIDVal, &genreName, &genreSlug,
		&planIDVal, &planName, &planTier,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("loading video: %w", err)
	}
	v.Slug = valueobject.SlugFromTrusted(slug)
	v.Duration = valueobject.NewDuration(durSec)
	v.UploadedAt = uploadedAt
	v.Type = entity.ContentTypeVideo
	if genreIDVal != nil {
		v.Genre = &entity.Genre{ID: *genreIDVal, Name: *genreName, Slug: *genreSlug}
	}
	if planIDVal != nil {
		v.Plan = &entity.Plan{ID: *planIDVal, Name: *planName, Tier: entity.PlanTier(*planTier)}
	}
	return &v, nil
}

// ─── Series ──────────────────────────────────────────────────────────────────

func (r *ContentRepository) CreateSeries(ctx context.Context, series *entity.Series) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO content (id, type, title, slug, description, synopsis,
		                     poster_url, backdrop_url, trailer_url, status, is_featured,
		                     created_at, updated_at)
		VALUES ($1, 'series', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, series.ID, series.Title, series.Slug.String(), series.Description, series.Synopsis,
		series.PosterURL, series.BackdropURL, series.TrailerURL,
		series.Status, series.IsFeatured, series.CreatedAt, series.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "content_slug_key" {
			return domain.ErrSlugAlreadyExists
		}
		return fmt.Errorf("creating series content row: %w", err)
	}
	return nil
}

func (r *ContentRepository) GetSeriesByID(ctx context.Context, id uuid.UUID) (*entity.Series, error) {
	return r.loadSeries(ctx, `WHERE c.id = $1`, id)
}

func (r *ContentRepository) GetSeriesBySlug(ctx context.Context, slug string) (*entity.Series, error) {
	return r.loadSeries(ctx, `WHERE c.slug = $1 AND c.status = 'published'`, slug)
}

func (r *ContentRepository) UpdateSeries(ctx context.Context, series *entity.Series) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE content SET title=$2, slug=$3, description=$4, synopsis=$5,
		       poster_url=$6, backdrop_url=$7, trailer_url=$8,
		       status=$9, is_featured=$10, updated_at=$11
		WHERE id = $1
	`, series.ID, series.Title, series.Slug.String(), series.Description, series.Synopsis,
		series.PosterURL, series.BackdropURL, series.TrailerURL,
		series.Status, series.IsFeatured, series.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating series: %w", err)
	}
	return nil
}

func (r *ContentRepository) loadSeries(ctx context.Context, where string, args ...any) (*entity.Series, error) {
	db := connFromCtx(ctx, r.pool)
	s := &entity.Series{}
	var slug string
	var totalSeasons int
	err := db.QueryRow(ctx, `
		SELECT c.id, c.title, c.slug, c.description, c.synopsis,
		       COALESCE(c.poster_url, ''), COALESCE(c.backdrop_url, ''), COALESCE(c.trailer_url, ''),
		       c.status, c.is_featured, c.created_at, c.updated_at,
		       COUNT(DISTINCT se.id) AS total_seasons
		FROM content c
		LEFT JOIN seasons se ON se.content_id = c.id
		`+where+`
		GROUP BY c.id`, args...).Scan(
		&s.ID, &s.Title, &slug, &s.Description, &s.Synopsis,
		&s.PosterURL, &s.BackdropURL, &s.TrailerURL,
		&s.Status, &s.IsFeatured, &s.CreatedAt, &s.UpdatedAt,
		&totalSeasons,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("loading series: %w", err)
	}
	s.Slug = valueobject.SlugFromTrusted(slug)
	s.TotalSeasons = totalSeasons
	return s, nil
}

// ─── Shared ──────────────────────────────────────────────────────────────────

func (r *ContentRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM content WHERE slug = $1)`, slug).Scan(&exists)
	return exists, err
}

func (r *ContentRepository) GetTypeByID(ctx context.Context, id uuid.UUID) (entity.ContentType, error) {
	db := connFromCtx(ctx, r.pool)
	var ct entity.ContentType
	err := db.QueryRow(ctx, `SELECT type FROM content WHERE id = $1`, id).Scan(&ct)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("getting content type: %w", err)
	}
	return ct, nil
}

func (r *ContentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM content WHERE id = $1`, id)
	return err
}

func (r *ContentRepository) SetStatus(ctx context.Context, id uuid.UUID, status entity.ContentStatus) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `UPDATE content SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	return err
}

// ─── List / Search ────────────────────────────────────────────────────────────

const summarySelect = `
	SELECT c.id, c.type, c.title, c.slug, c.description, c.rating,
	       COALESCE(c.poster_url, ''), c.status, c.is_featured, c.created_at, c.updated_at,
	       c.genre_id, g.name, g.slug,
	       c.plan_id, p.name, p.tier,
	       COALESCE(v.creator_name, '')
	FROM content c
	LEFT JOIN genres g ON g.id = c.genre_id
	LEFT JOIN plans  p ON p.id = c.plan_id
	LEFT JOIN videos v ON v.content_id = c.id AND c.type = 'video'
`

func scanSummaryRow(rows pgx.Rows) (entity.ContentSummary, error) {
	var (
		s          entity.ContentSummary
		slug       string
		genreIDVal *uuid.UUID
		genreName  *string
		genreSlug  *string
		planIDVal  *uuid.UUID
		planName   *string
		planTier   *string
	)
	err := rows.Scan(
		&s.ID, &s.Type, &s.Title, &slug, &s.Description, &s.Rating,
		&s.PosterURL, &s.Status, &s.IsFeatured, &s.CreatedAt, &s.UpdatedAt,
		&genreIDVal, &genreName, &genreSlug,
		&planIDVal, &planName, &planTier,
		&s.CreatorName,
	)
	if err != nil {
		return s, err
	}
	s.Slug = valueobject.SlugFromTrusted(slug)
	if genreIDVal != nil {
		s.Genre = &entity.Genre{ID: *genreIDVal, Name: *genreName, Slug: *genreSlug}
	}
	if planIDVal != nil {
		s.Plan = &entity.Plan{ID: *planIDVal, Name: *planName, Tier: entity.PlanTier(*planTier)}
	}
	return s, nil
}

func (r *ContentRepository) Search(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.ContentSummary, int64, error) {
	db := connFromCtx(ctx, r.pool)

	whereClause, countArgs, nextIdx := filter.CountSQL(fs, "c.status = 'published'", 1)
	extraWhere, searchArgs := searchCondition(searchQuery, nextIdx)

	var total int64
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM content c `+whereClause+extraWhere,
		append(countArgs, searchArgs...)...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting content: %w", err)
	}

	sql := filter.ToSQL(fs, "c.status = 'published'", 1)
	dataExtraWhere, dataSearchArgs := searchCondition(searchQuery, sql.NextArgIdx)

	rows, err := db.Query(ctx,
		summarySelect+sql.WhereClause+dataExtraWhere+" "+sql.OrderClause+" "+sql.LimitClause,
		append(sql.Args, dataSearchArgs...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("searching content: %w", err)
	}
	defer rows.Close()
	summaries, err := collectSummaries(rows)
	return summaries, total, err
}

func (r *ContentRepository) AdminSearch(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.ContentSummary, int64, error) {
	db := connFromCtx(ctx, r.pool)

	whereClause, countArgs, nextIdx := filter.CountSQL(fs, "", 1)
	extraWhere, searchArgs := searchConditionAdmin(searchQuery, nextIdx, whereClause == "")

	var total int64
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM content c `+whereClause+extraWhere,
		append(countArgs, searchArgs...)...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting admin content: %w", err)
	}

	sql := filter.ToSQL(fs, "", 1)
	dataExtraWhere, dataSearchArgs := searchConditionAdmin(searchQuery, sql.NextArgIdx, sql.WhereClause == "")

	rows, err := db.Query(ctx,
		summarySelect+sql.WhereClause+dataExtraWhere+" "+sql.OrderClause+" "+sql.LimitClause,
		append(sql.Args, dataSearchArgs...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("admin searching content: %w", err)
	}
	defer rows.Close()
	summaries, err := collectSummaries(rows)
	return summaries, total, err
}

func (r *ContentRepository) GetFeatured(ctx context.Context, limit int) ([]entity.ContentSummary, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx,
		summarySelect+`WHERE c.status = 'published' AND c.is_featured = true
		ORDER BY c.updated_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("getting featured content: %w", err)
	}
	defer rows.Close()
	return collectSummaries(rows)
}

func (r *ContentRepository) GetByGenre(ctx context.Context, genreID uuid.UUID, page, perPage int) ([]entity.ContentSummary, int64, error) {
	db := connFromCtx(ctx, r.pool)
	offset := (page - 1) * perPage

	var total int64
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM content c WHERE c.genre_id = $1 AND c.status = 'published'`,
		genreID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting genre content: %w", err)
	}

	rows, err := db.Query(ctx,
		summarySelect+`WHERE c.genre_id = $1 AND c.status = 'published'
		ORDER BY c.created_at DESC LIMIT $2 OFFSET $3`,
		genreID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("getting genre content: %w", err)
	}
	defer rows.Close()
	summaries, err := collectSummaries(rows)
	return summaries, total, err
}

// ─── Seasons / Episodes ───────────────────────────────────────────────────────

func (r *ContentRepository) GetSeasonsForSeries(ctx context.Context, seriesID uuid.UUID) ([]entity.Season, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, content_id, number, title,
		       COALESCE(trailer_url, ''), COALESCE(backdrop_url, ''), created_at
		FROM seasons WHERE content_id = $1 ORDER BY number
	`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("getting seasons: %w", err)
	}
	defer rows.Close()

	var seasons []entity.Season
	for rows.Next() {
		var s entity.Season
		if err := rows.Scan(&s.ID, &s.SeriesID, &s.Number, &s.Title,
			&s.TrailerURL, &s.BackdropURL, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning season: %w", err)
		}
		seasons = append(seasons, s)
	}
	return seasons, nil
}

func (r *ContentRepository) GetEpisodesForSeason(ctx context.Context, seasonID uuid.UUID) ([]entity.Episode, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, season_id, COALESCE(series_id, '00000000-0000-0000-0000-000000000000'::uuid),
		       number, title, synopsis, duration_seconds, air_date, created_at,
		       COALESCE(director, ''), '[]'::jsonb
		FROM episodes WHERE season_id = $1 ORDER BY number
	`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("getting episodes: %w", err)
	}
	defer rows.Close()
	return collectEpisodes(rows)
}

func (r *ContentRepository) GetEpisodeByID(ctx context.Context, id uuid.UUID) (*entity.Episode, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, season_id, COALESCE(series_id, '00000000-0000-0000-0000-000000000000'::uuid),
		       number, title, synopsis, duration_seconds, air_date, created_at,
		       COALESCE(director, ''), '[]'::jsonb
		FROM episodes WHERE id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("getting episode: %w", err)
	}
	defer rows.Close()
	episodes, err := collectEpisodes(rows)
	if err != nil {
		return nil, err
	}
	if len(episodes) == 0 {
		return nil, nil
	}
	return &episodes[0], nil
}

// ─── Video Assets ─────────────────────────────────────────────────────────────

func (r *ContentRepository) CreateVideoAsset(ctx context.Context, asset *entity.VideoAsset, contentID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO video_assets (id, content_id, storage_key, status, qualities, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())
		ON CONFLICT (content_id) WHERE episode_id IS NULL
		DO UPDATE SET storage_key = EXCLUDED.storage_key,
		              status      = EXCLUDED.status,
		              qualities   = EXCLUDED.qualities,
		              updated_at  = now()
	`, asset.ID, contentID, asset.StorageKey, asset.Status, "[]")
	return err
}

func (r *ContentRepository) GetVideoAssetForContent(ctx context.Context, contentID uuid.UUID) (*entity.VideoAsset, error) {
	db := connFromCtx(ctx, r.pool)
	a := &entity.VideoAsset{}
	var qualitiesJSON []byte
	err := db.QueryRow(ctx, `
		SELECT id, storage_key, status, qualities
		FROM video_assets WHERE content_id = $1 AND episode_id IS NULL
	`, contentID).Scan(&a.ID, &a.StorageKey, &a.Status, &qualitiesJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting video asset: %w", err)
	}
	if len(qualitiesJSON) > 0 {
		_ = json.Unmarshal(qualitiesJSON, &a.Qualities)
	}
	return a, nil
}

func (r *ContentRepository) UpdateVideoAssetStatus(ctx context.Context, assetID uuid.UUID, status entity.VideoAssetStatus) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE video_assets SET status = $2, updated_at = now() WHERE id = $1
	`, assetID, status)
	return err
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func genreID(g *entity.Genre) *uuid.UUID {
	if g == nil {
		return nil
	}
	return &g.ID
}

func planID(p *entity.Plan) *uuid.UUID {
	if p == nil {
		return nil
	}
	return &p.ID
}

func searchCondition(q string, nextIdx int) (string, []any) {
	if q == "" {
		return "", nil
	}
	return fmt.Sprintf(" AND (c.title ILIKE $%d OR c.description ILIKE $%d)", nextIdx, nextIdx),
		[]any{"%" + q + "%"}
}

func searchConditionAdmin(q string, nextIdx int, noWhere bool) (string, []any) {
	if q == "" {
		return "", nil
	}
	keyword := "AND"
	if noWhere {
		keyword = "WHERE"
	}
	return fmt.Sprintf(" %s (c.title ILIKE $%d OR c.description ILIKE $%d)", keyword, nextIdx, nextIdx),
		[]any{"%" + q + "%"}
}

func collectSummaries(rows pgx.Rows) ([]entity.ContentSummary, error) {
	var summaries []entity.ContentSummary
	for rows.Next() {
		s, err := scanSummaryRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning content summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

func collectEpisodes(rows pgx.Rows) ([]entity.Episode, error) {
	var episodes []entity.Episode
	for rows.Next() {
		var e entity.Episode
		var durSec int64
		var castJSON []byte
		if err := rows.Scan(&e.ID, &e.SeasonID, &e.SeriesID, &e.Number, &e.Title, &e.Synopsis,
			&durSec, &e.AirDate, &e.CreatedAt, &e.Director, &castJSON); err != nil {
			return nil, fmt.Errorf("scanning episode: %w", err)
		}
		e.Duration = valueobject.NewDuration(durSec)
		if len(castJSON) > 0 {
			_ = json.Unmarshal(castJSON, &e.CastMembers)
		}
		episodes = append(episodes, e)
	}
	return episodes, nil
}
