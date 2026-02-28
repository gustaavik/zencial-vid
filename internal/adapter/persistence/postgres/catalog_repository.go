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

// CatalogRepository implements repository.CatalogRepository using PostgreSQL.
type CatalogRepository struct {
	pool *pgxpool.Pool
}

// NewCatalogRepository creates a new CatalogRepository.
func NewCatalogRepository(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

func (r *CatalogRepository) CreateGenre(ctx context.Context, genre *entity.Genre) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `INSERT INTO genres (id, name, slug) VALUES ($1, $2, $3)`,
		genre.ID, genre.Name, genre.Slug)
	if err != nil {
		return fmt.Errorf("creating genre: %w", err)
	}
	return nil
}

func (r *CatalogRepository) GetGenreByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error) {
	db := connFromCtx(ctx, r.pool)
	g := &entity.Genre{}
	err := db.QueryRow(ctx, `SELECT id, name, slug FROM genres WHERE id = $1`, id).
		Scan(&g.ID, &g.Name, &g.Slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting genre by id: %w", err)
	}
	return g, nil
}

func (r *CatalogRepository) GetGenreBySlug(ctx context.Context, slug string) (*entity.Genre, error) {
	db := connFromCtx(ctx, r.pool)
	g := &entity.Genre{}
	err := db.QueryRow(ctx, `SELECT id, name, slug FROM genres WHERE slug = $1`, slug).
		Scan(&g.ID, &g.Name, &g.Slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting genre by slug: %w", err)
	}
	return g, nil
}

func (r *CatalogRepository) ListGenres(ctx context.Context) ([]entity.Genre, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `SELECT id, name, slug FROM genres ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing genres: %w", err)
	}
	defer rows.Close()

	var genres []entity.Genre
	for rows.Next() {
		var g entity.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.Slug); err != nil {
			return nil, fmt.Errorf("scanning genre: %w", err)
		}
		genres = append(genres, g)
	}
	return genres, nil
}

func (r *CatalogRepository) UpdateGenre(ctx context.Context, genre *entity.Genre) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `UPDATE genres SET name = $2, slug = $3 WHERE id = $1`,
		genre.ID, genre.Name, genre.Slug)
	if err != nil {
		return fmt.Errorf("updating genre: %w", err)
	}
	return nil
}

func (r *CatalogRepository) DeleteGenre(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM genres WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting genre: %w", err)
	}
	return nil
}

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]entity.Category, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `SELECT id, name, slug, description, parent_id, sort_order FROM categories ORDER BY sort_order, name`)
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.ParentID, &c.SortOrder); err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *CatalogRepository) GetCategoryBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	db := connFromCtx(ctx, r.pool)
	c := &entity.Category{}
	err := db.QueryRow(ctx, `SELECT id, name, slug, description, parent_id, sort_order FROM categories WHERE slug = $1`, slug).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.ParentID, &c.SortOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting category by slug: %w", err)
	}
	return c, nil
}
