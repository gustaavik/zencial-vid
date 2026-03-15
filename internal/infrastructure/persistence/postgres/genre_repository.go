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

var genreFilterConfig = filter.Config{
	Columns: map[string]filter.ColumnDef{
		"name":          {DBColumn: "gt.name", AllowedOps: []filter.Op{filter.OpLike}, Type: filter.TypeString},
		"language_code": {DBColumn: "gt.language_code", AllowedOps: []filter.Op{filter.OpEq}, Type: filter.TypeString},
	},
	SortColumns: map[string]filter.SortDef{
		"name":       {DBColumn: "gt.name"},
		"created_at": {DBColumn: "g.created_at"},
	},
	DefaultSort: "g.created_at DESC",
}

// GenreFilterConfig returns the filter configuration for genres.
func GenreFilterConfig() filter.Config {
	return genreFilterConfig
}

// GenreRepository implements repository.GenreRepository using PostgreSQL.
type GenreRepository struct {
	pool *pgxpool.Pool
}

// NewGenreRepository creates a new GenreRepository.
func NewGenreRepository(pool *pgxpool.Pool) *GenreRepository {
	return &GenreRepository{pool: pool}
}

func (r *GenreRepository) Create(ctx context.Context, genre *entity.Genre) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		INSERT INTO genres (id, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`, genre.ID, genre.Slug.String(), genre.CreatedAt, genre.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating genre: %w", err)
	}

	for _, t := range genre.Translations {
		_, err := db.Exec(ctx, `
			INSERT INTO genre_translations (id, genre_id, language_code, name, description)
			VALUES ($1, $2, $3, $4, $5)
		`, t.ID, genre.ID, t.LanguageCode.String(), t.Name, t.Description)
		if err != nil {
			return fmt.Errorf("creating genre translation: %w", err)
		}
	}

	return nil
}

func (r *GenreRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error) {
	db := connFromCtx(ctx, r.pool)
	genre := &entity.Genre{}

	var slug string

	err := db.QueryRow(ctx, `
		SELECT id, slug, created_at, updated_at FROM genres WHERE id = $1
	`, id).Scan(&genre.ID, &slug, &genre.CreatedAt, &genre.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting genre by id: %w", err)
	}

	translations, err := r.getTranslations(ctx, db, genre.ID)
	if err != nil {
		return nil, err
	}
	genre.Translations = translations
	genre.Slug = valueobject.SlugFromTrusted(slug)

	return genre, nil
}

func (r *GenreRepository) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Genre, error) {
	db := connFromCtx(ctx, r.pool)
	genre := &entity.Genre{}

	var slugStr string

	err := db.QueryRow(ctx, `
		SELECT id, slug, created_at, updated_at FROM genres WHERE slug = $1
	`, slug.String()).Scan(&genre.ID, &slugStr, &genre.CreatedAt, &genre.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting genre by slug: %w", err)
	}

	translations, err := r.getTranslations(ctx, db, genre.ID)
	if err != nil {
		return nil, err
	}
	genre.Translations = translations
	genre.Slug = valueobject.SlugFromTrusted(slugStr)

	return genre, nil
}

func (r *GenreRepository) Update(ctx context.Context, genre *entity.Genre) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		UPDATE genres SET slug = $2, updated_at = $3 WHERE id = $1
	`, genre.ID, genre.Slug.String(), genre.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating genre: %w", err)
	}

	// Delete existing translations and re-insert
	_, err = db.Exec(ctx, `DELETE FROM genre_translations WHERE genre_id = $1`, genre.ID)
	if err != nil {
		return fmt.Errorf("deleting genre translations: %w", err)
	}

	for _, t := range genre.Translations {
		_, err := db.Exec(ctx, `
			INSERT INTO genre_translations (id, genre_id, language_code, name, description)
			VALUES ($1, $2, $3, $4, $5)
		`, t.ID, genre.ID, t.LanguageCode.String(), t.Name, t.Description)
		if err != nil {
			return fmt.Errorf("inserting genre translation: %w", err)
		}
	}

	return nil
}

func (r *GenreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM genres WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting genre: %w", err)
	}
	return nil
}

func (r *GenreRepository) List(ctx context.Context, fs filter.FilterSet) ([]entity.Genre, int64, error) {
	db := connFromCtx(ctx, r.pool)

	// Count query
	countWhere, countArgs, _ := filter.CountSQL(fs, "", 1)
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT g.id)
		FROM genres g
		LEFT JOIN genre_translations gt ON g.id = gt.genre_id
		%s
	`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting genres: %w", err)
	}

	// Data query
	sql := filter.ToSQL(fs, "", 1)
	dataQuery := fmt.Sprintf(`
		SELECT DISTINCT g.id, g.slug, g.created_at, g.updated_at
		FROM genres g
		LEFT JOIN genre_translations gt ON g.id = gt.genre_id
		%s %s %s
	`, sql.WhereClause, sql.OrderClause, sql.LimitClause)

	rows, err := db.Query(ctx, dataQuery, sql.Args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing genres: %w", err)
	}
	defer rows.Close()

	var genres []entity.Genre
	for rows.Next() {
		var g entity.Genre
		var slug string

		if err := rows.Scan(&g.ID, &slug, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning genre: %w", err)
		}
		g.Slug = valueobject.SlugFromTrusted(slug)
		genres = append(genres, g)
	}

	// Load translations for each genre
	for i := range genres {
		translations, err := r.getTranslations(ctx, db, genres[i].ID)
		if err != nil {
			return nil, 0, err
		}
		genres[i].Translations = translations
	}

	return genres, total, nil
}

func (r *GenreRepository) ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM genres WHERE slug = $1)`, slug.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking slug existence: %w", err)
	}
	return exists, nil
}

func (r *GenreRepository) getTranslations(ctx context.Context, db DBTX, genreID uuid.UUID) ([]entity.GenreTranslation, error) {
	rows, err := db.Query(ctx, `
		SELECT id, genre_id, language_code, name, description
		FROM genre_translations
		WHERE genre_id = $1
	`, genreID)
	if err != nil {
		return nil, fmt.Errorf("getting genre translations: %w", err)
	}
	defer rows.Close()

	var translations []entity.GenreTranslation
	for rows.Next() {
		var t entity.GenreTranslation
		var langCode string
		if err := rows.Scan(&t.ID, &t.GenreID, &langCode, &t.Name, &t.Description); err != nil {
			return nil, fmt.Errorf("scanning genre translation: %w", err)
		}
		t.LanguageCode = valueobject.LanguageCodeFromTrusted(langCode)
		translations = append(translations, t)
	}

	return translations, nil
}
