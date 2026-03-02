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

// ContentRepository implements repository.ContentRepository using PostgreSQL.
type ContentRepository struct {
	pool *pgxpool.Pool
}

// NewContentRepository creates a new ContentRepository.
func NewContentRepository(pool *pgxpool.Pool) *ContentRepository {
	return &ContentRepository{pool: pool}
}

func (r *ContentRepository) Create(ctx context.Context, content *entity.Content) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO content (id, type, title, slug, description, synopsis, rating, release_year,
		                     poster_url, backdrop_url, trailer_url, director, status, is_featured, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, content.ID, content.Type, content.Title, content.Slug.String(), content.Description,
		content.Synopsis, content.Rating, content.ReleaseYear, content.PosterURL,
		content.BackdropURL, content.TrailerURL, content.Director, content.Status,
		content.IsFeatured, content.CreatedAt, content.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating content: %w", err)
	}
	return nil
}

func (r *ContentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Content, error) {
	db := connFromCtx(ctx, r.pool)
	c := &entity.Content{}
	var slug string
	err := db.QueryRow(ctx, `
		SELECT id, type, title, slug, description, synopsis, rating, release_year,
		       poster_url, backdrop_url, trailer_url, director, status, is_featured, created_at, updated_at
		FROM content WHERE id = $1
	`, id).Scan(&c.ID, &c.Type, &c.Title, &slug, &c.Description, &c.Synopsis,
		&c.Rating, &c.ReleaseYear, &c.PosterURL, &c.BackdropURL, &c.TrailerURL,
		&c.Director, &c.Status, &c.IsFeatured, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting content by id: %w", err)
	}
	c.Slug = valueobject.SlugFromTrusted(slug)
	return c, nil
}

func (r *ContentRepository) GetBySlug(ctx context.Context, slug string) (*entity.Content, error) {
	db := connFromCtx(ctx, r.pool)
	c := &entity.Content{}
	var slugStr string
	err := db.QueryRow(ctx, `
		SELECT id, type, title, slug, description, synopsis, rating, release_year,
		       poster_url, backdrop_url, trailer_url, director, status, is_featured, created_at, updated_at
		FROM content WHERE slug = $1 AND status = 'published'
	`, slug).Scan(&c.ID, &c.Type, &c.Title, &slugStr, &c.Description, &c.Synopsis,
		&c.Rating, &c.ReleaseYear, &c.PosterURL, &c.BackdropURL, &c.TrailerURL,
		&c.Director, &c.Status, &c.IsFeatured, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting content by slug: %w", err)
	}
	c.Slug = valueobject.SlugFromTrusted(slugStr)
	return c, nil
}

func (r *ContentRepository) Update(ctx context.Context, content *entity.Content) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE content SET title=$2, slug=$3, description=$4, synopsis=$5, rating=$6,
		       release_year=$7, poster_url=$8, backdrop_url=$9, trailer_url=$10,
		       director=$11, status=$12, is_featured=$13, updated_at=$14
		WHERE id = $1
	`, content.ID, content.Title, content.Slug.String(), content.Description,
		content.Synopsis, content.Rating, content.ReleaseYear, content.PosterURL,
		content.BackdropURL, content.TrailerURL, content.Director, content.Status,
		content.IsFeatured, content.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating content: %w", err)
	}
	return nil
}

func (r *ContentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM content WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting content: %w", err)
	}
	return nil
}

const contentBaseCondition = "c.status = 'published'"

func (r *ContentRepository) Search(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.Content, int64, error) {
	db := connFromCtx(ctx, r.pool)

	// Build count query
	whereClause, countArgs, nextIdx := filter.CountSQL(fs, contentBaseCondition, 1)

	// Append free-text search condition (uses OR across multiple columns)
	extraWhere := ""
	var searchArgs []interface{}
	if searchQuery != "" {
		extraWhere = fmt.Sprintf(" AND (c.title ILIKE $%d OR c.description ILIKE $%d)", nextIdx, nextIdx)
		searchArgs = append(searchArgs, "%"+searchQuery+"%")
	}

	countAllArgs := append(countArgs, searchArgs...)
	countQ := `SELECT COUNT(*) FROM content c ` + whereClause + extraWhere
	var total int64
	if err := db.QueryRow(ctx, countQ, countAllArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting content: %w", err)
	}

	// Build data query
	sql := filter.ToSQL(fs, contentBaseCondition, 1)

	// Re-apply search condition for data query
	dataExtraWhere := ""
	var dataSearchArgs []interface{}
	if searchQuery != "" {
		dataExtraWhere = fmt.Sprintf(" AND (c.title ILIKE $%d OR c.description ILIKE $%d)", sql.NextArgIdx, sql.NextArgIdx)
		dataSearchArgs = append(dataSearchArgs, "%"+searchQuery+"%")
	}

	dataQ := `SELECT c.id, c.type, c.title, c.slug, c.description, c.rating, c.release_year,
	                 c.poster_url, c.status, c.is_featured, c.created_at, c.updated_at
	          FROM content c ` + sql.WhereClause + dataExtraWhere + ` ` + sql.OrderClause + ` ` + sql.LimitClause
	dataAllArgs := append(sql.Args, dataSearchArgs...)

	rows, err := db.Query(ctx, dataQ, dataAllArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("searching content: %w", err)
	}
	defer rows.Close()

	var contents []entity.Content
	for rows.Next() {
		var c entity.Content
		var slug string
		err := rows.Scan(&c.ID, &c.Type, &c.Title, &slug, &c.Description,
			&c.Rating, &c.ReleaseYear, &c.PosterURL, &c.Status, &c.IsFeatured,
			&c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning content: %w", err)
		}
		c.Slug = valueobject.SlugFromTrusted(slug)
		contents = append(contents, c)
	}
	return contents, total, nil
}

func (r *ContentRepository) GetFeatured(ctx context.Context, limit int) ([]entity.Content, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, type, title, slug, description, rating, release_year, poster_url, status, is_featured, created_at, updated_at
		FROM content WHERE status = 'published' AND is_featured = true
		ORDER BY updated_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("getting featured content: %w", err)
	}
	defer rows.Close()

	var contents []entity.Content
	for rows.Next() {
		var c entity.Content
		var slug string
		if err := rows.Scan(&c.ID, &c.Type, &c.Title, &slug, &c.Description,
			&c.Rating, &c.ReleaseYear, &c.PosterURL, &c.Status, &c.IsFeatured,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning featured content: %w", err)
		}
		c.Slug = valueobject.SlugFromTrusted(slug)
		contents = append(contents, c)
	}
	return contents, nil
}

func (r *ContentRepository) GetByGenre(ctx context.Context, genreID uuid.UUID, page, perPage int) ([]entity.Content, int64, error) {
	db := connFromCtx(ctx, r.pool)
	offset := (page - 1) * perPage

	var total int64
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM content c
		JOIN content_genres cg ON c.id = cg.content_id
		WHERE cg.genre_id = $1 AND c.status = 'published'
	`, genreID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting genre content: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT c.id, c.type, c.title, c.slug, c.description, c.rating, c.release_year,
		       c.poster_url, c.status, c.is_featured, c.created_at, c.updated_at
		FROM content c
		JOIN content_genres cg ON c.id = cg.content_id
		WHERE cg.genre_id = $1 AND c.status = 'published'
		ORDER BY c.created_at DESC LIMIT $2 OFFSET $3
	`, genreID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("getting genre content: %w", err)
	}
	defer rows.Close()

	var contents []entity.Content
	for rows.Next() {
		var c entity.Content
		var slug string
		if err := rows.Scan(&c.ID, &c.Type, &c.Title, &slug, &c.Description,
			&c.Rating, &c.ReleaseYear, &c.PosterURL, &c.Status, &c.IsFeatured,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning genre content: %w", err)
		}
		c.Slug = valueobject.SlugFromTrusted(slug)
		contents = append(contents, c)
	}
	return contents, total, nil
}

func (r *ContentRepository) GetSeasonsForContent(ctx context.Context, contentID uuid.UUID) ([]entity.Season, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, content_id, number, title, created_at
		FROM seasons WHERE content_id = $1 ORDER BY number
	`, contentID)
	if err != nil {
		return nil, fmt.Errorf("getting seasons: %w", err)
	}
	defer rows.Close()

	var seasons []entity.Season
	for rows.Next() {
		var s entity.Season
		if err := rows.Scan(&s.ID, &s.ContentID, &s.Number, &s.Title, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning season: %w", err)
		}
		seasons = append(seasons, s)
	}
	return seasons, nil
}

func (r *ContentRepository) GetEpisodesForSeason(ctx context.Context, seasonID uuid.UUID) ([]entity.Episode, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, season_id, number, title, synopsis, duration_seconds, air_date, created_at
		FROM episodes WHERE season_id = $1 ORDER BY number
	`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("getting episodes: %w", err)
	}
	defer rows.Close()

	var episodes []entity.Episode
	for rows.Next() {
		var e entity.Episode
		var durationSeconds int64
		if err := rows.Scan(&e.ID, &e.SeasonID, &e.Number, &e.Title, &e.Synopsis,
			&durationSeconds, &e.AirDate, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning episode: %w", err)
		}
		e.Duration = valueobject.NewDuration(durationSeconds)
		episodes = append(episodes, e)
	}
	return episodes, nil
}

func (r *ContentRepository) GetEpisodeByID(ctx context.Context, id uuid.UUID) (*entity.Episode, error) {
	db := connFromCtx(ctx, r.pool)
	e := &entity.Episode{}
	var durationSeconds int64
	err := db.QueryRow(ctx, `
		SELECT id, season_id, number, title, synopsis, duration_seconds, air_date, created_at
		FROM episodes WHERE id = $1
	`, id).Scan(&e.ID, &e.SeasonID, &e.Number, &e.Title, &e.Synopsis,
		&durationSeconds, &e.AirDate, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting episode: %w", err)
	}
	e.Duration = valueobject.NewDuration(durationSeconds)
	return e, nil
}
