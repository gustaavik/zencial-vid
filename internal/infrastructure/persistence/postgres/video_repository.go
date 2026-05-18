package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

var videoFilterConfig = filter.Config{
	Columns: map[string]filter.ColumnDef{
		"status":         {DBColumn: "v.status", AllowedOps: []filter.Op{filter.OpEq}, Type: filter.TypeString},
		"creator":        {DBColumn: "v.creator", AllowedOps: []filter.Op{filter.OpEq, filter.OpLike}, Type: filter.TypeString},
		"content_rating": {DBColumn: "v.content_rating", AllowedOps: []filter.Op{filter.OpEq, filter.OpIn}, Type: filter.TypeString},
		"title":          {DBColumn: "v.title", AllowedOps: []filter.Op{filter.OpLike}, Type: filter.TypeString},
	},
	SortColumns: map[string]filter.SortDef{
		"title":      {DBColumn: "v.title"},
		"created_at": {DBColumn: "v.created_at"},
		"duration":   {DBColumn: "v.duration"},
	},
	DefaultSort: "v.created_at DESC",
}

// VideoFilterConfig returns the filter configuration for videos.
func VideoFilterConfig() filter.Config {
	return videoFilterConfig
}

// VideoRepository implements repository.VideoRepository using PostgreSQL.
type VideoRepository struct {
	pool *pgxpool.Pool
}

// NewVideoRepository creates a new VideoRepository.
func NewVideoRepository(pool *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{pool: pool}
}

func (r *VideoRepository) Create(ctx context.Context, video *entity.Video) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		INSERT INTO videos (id, title, slug, description, creator, duration, content_rating,
		                    status, storage_key, content_type, file_size, thumbnail_key, uploaded_by,
		                    minimum_plan_level, transcode_error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`, video.ID, video.Title, video.Slug.String(), video.Description, video.Creator,
		video.Duration.Seconds, video.ContentRating,
		string(video.Status), video.StorageKey, video.ContentType, video.FileSize,
		video.ThumbnailKey, video.UploadedBy, video.MinimumPlanLevel, video.TranscodeError,
		video.CreatedAt, video.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating video: %w", err)
	}

	if len(video.GenreIDs) > 0 {
		if err := r.SetGenres(ctx, video.ID, video.GenreIDs); err != nil {
			return err
		}
	}

	return nil
}

func (r *VideoRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	db := connFromCtx(ctx, r.pool)
	return r.scanVideo(ctx, db, `
		SELECT id, title, slug, description, creator, duration, content_rating,
		       status, storage_key, content_type, file_size, thumbnail_key, uploaded_by,
		       minimum_plan_level, transcode_error, created_at, updated_at
		FROM videos WHERE id = $1
	`, id)
}

func (r *VideoRepository) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Video, error) {
	db := connFromCtx(ctx, r.pool)
	return r.scanVideo(ctx, db, `
		SELECT id, title, slug, description, creator, duration, content_rating,
		       status, storage_key, content_type, file_size, thumbnail_key, uploaded_by,
		       minimum_plan_level, transcode_error, created_at, updated_at
		FROM videos WHERE slug = $1
	`, slug.String())
}

func (r *VideoRepository) Update(ctx context.Context, video *entity.Video) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		UPDATE videos SET title = $2, slug = $3, description = $4, creator = $5,
		       duration = $6, content_rating = $7, status = $8,
		       storage_key = $9, content_type = $10, file_size = $11, thumbnail_key = $12,
		       minimum_plan_level = $13, transcode_error = $14, updated_at = $15
		WHERE id = $1
	`, video.ID, video.Title, video.Slug.String(), video.Description, video.Creator,
		video.Duration.Seconds, video.ContentRating,
		string(video.Status), video.StorageKey, video.ContentType, video.FileSize,
		video.ThumbnailKey, video.MinimumPlanLevel, video.TranscodeError, video.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating video: %w", err)
	}

	return nil
}

func (r *VideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM videos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting video: %w", err)
	}
	return nil
}

func (r *VideoRepository) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	return r.listWithBase(ctx, fs, "")
}

func (r *VideoRepository) ListPublished(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	return r.listWithBase(ctx, fs, "v.status = 'published'")
}

func (r *VideoRepository) ListByUploader(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	db := connFromCtx(ctx, r.pool)

	// uploaderID is $1; filter conditions start at $2.
	baseCondition := "v.uploaded_by = $1"
	baseArgs := []any{uploaderID}

	countWhere, countFilterArgs, _ := filter.CountSQL(fs, baseCondition, 2)
	countArgs := append(append([]any{}, baseArgs...), countFilterArgs...)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM videos v %s`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting videos by uploader: %w", err)
	}

	sql := filter.ToSQL(fs, baseCondition, 2)
	dataArgs := append(append([]any{}, baseArgs...), sql.Args...)
	dataQuery := fmt.Sprintf(`
		SELECT id, title, slug, description, creator, duration, content_rating,
		       status, storage_key, content_type, file_size, thumbnail_key, uploaded_by,
		       minimum_plan_level, transcode_error, created_at, updated_at
		FROM videos v
		%s %s %s
	`, sql.WhereClause, sql.OrderClause, sql.LimitClause)

	rows, err := db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing videos by uploader: %w", err)
	}
	defer rows.Close()

	var videos []entity.Video
	for rows.Next() {
		v, err := r.scanVideoRow(rows)
		if err != nil {
			return nil, 0, err
		}
		videos = append(videos, *v)
	}

	for i := range videos {
		genreIDs, err := r.GetGenreIDs(ctx, videos[i].ID)
		if err != nil {
			return nil, 0, err
		}
		videos[i].GenreIDs = genreIDs
	}

	return videos, total, nil
}

func (r *VideoRepository) listWithBase(ctx context.Context, fs *filter.FilterSet, baseCondition string) ([]entity.Video, int64, error) {
	db := connFromCtx(ctx, r.pool)

	// Count
	countWhere, countArgs, _ := filter.CountSQL(fs, baseCondition, 1)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM videos v %s`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting videos: %w", err)
	}

	// Data
	sql := filter.ToSQL(fs, baseCondition, 1)
	dataQuery := fmt.Sprintf(`
		SELECT id, title, slug, description, creator, duration, content_rating,
		       status, storage_key, content_type, file_size, thumbnail_key, uploaded_by,
		       minimum_plan_level, transcode_error, created_at, updated_at
		FROM videos v
		%s %s %s
	`, sql.WhereClause, sql.OrderClause, sql.LimitClause)

	rows, err := db.Query(ctx, dataQuery, sql.Args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing videos: %w", err)
	}
	defer rows.Close()

	var videos []entity.Video
	for rows.Next() {
		v, err := r.scanVideoRow(rows)
		if err != nil {
			return nil, 0, err
		}
		videos = append(videos, *v)
	}

	// Load genre IDs for each video
	for i := range videos {
		genreIDs, err := r.GetGenreIDs(ctx, videos[i].ID)
		if err != nil {
			return nil, 0, err
		}
		videos[i].GenreIDs = genreIDs
	}

	return videos, total, nil
}

func (r *VideoRepository) ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM videos WHERE slug = $1)`, slug.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking video slug existence: %w", err)
	}
	return exists, nil
}

func (r *VideoRepository) SetGenres(ctx context.Context, videoID uuid.UUID, genreIDs []uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `DELETE FROM video_genres WHERE video_id = $1`, videoID)
	if err != nil {
		return fmt.Errorf("clearing video genres: %w", err)
	}

	for _, genreID := range genreIDs {
		_, err := db.Exec(ctx, `
			INSERT INTO video_genres (video_id, genre_id) VALUES ($1, $2)
		`, videoID, genreID)
		if err != nil {
			return fmt.Errorf("setting video genre: %w", err)
		}
	}

	return nil
}

func (r *VideoRepository) GetGenreIDs(ctx context.Context, videoID uuid.UUID) ([]uuid.UUID, error) {
	db := connFromCtx(ctx, r.pool)

	rows, err := db.Query(ctx, `SELECT genre_id FROM video_genres WHERE video_id = $1`, videoID)
	if err != nil {
		return nil, fmt.Errorf("getting video genre IDs: %w", err)
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

func (r *VideoRepository) ListAllStorageKeys(ctx context.Context) ([]repository.VideoStorageInfo, error) {
	db := connFromCtx(ctx, r.pool)

	rows, err := db.Query(ctx, `SELECT id, storage_key, COALESCE(thumbnail_key, '') FROM videos`)
	if err != nil {
		return nil, fmt.Errorf("listing video storage keys: %w", err)
	}
	defer rows.Close()

	var infos []repository.VideoStorageInfo
	for rows.Next() {
		var info repository.VideoStorageInfo
		if err := rows.Scan(&info.ID, &info.StorageKey, &info.ThumbnailKey); err != nil {
			return nil, fmt.Errorf("scanning video storage info: %w", err)
		}
		infos = append(infos, info)
	}

	return infos, nil
}

func (r *VideoRepository) scanVideo(ctx context.Context, db DBTX, query string, args ...any) (*entity.Video, error) {
	var v entity.Video
	var slug, contentRating, status, transcodeError string
	var duration int64

	err := db.QueryRow(ctx, query, args...).Scan(
		&v.ID, &v.Title, &slug, &v.Description, &v.Creator,
		&duration, &contentRating, &status,
		&v.StorageKey, &v.ContentType, &v.FileSize, &v.ThumbnailKey,
		&v.UploadedBy, &v.MinimumPlanLevel, &transcodeError, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning video: %w", err)
	}

	v.Slug = valueobject.SlugFromTrusted(slug)
	v.Duration = valueobject.NewDuration(duration)
	v.ContentRating = contentRating
	v.Status = entity.VideoStatus(status)
	v.TranscodeError = transcodeError

	// Load genre IDs
	genreIDs, err := r.GetGenreIDs(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	v.GenreIDs = genreIDs

	return &v, nil
}

func (r *VideoRepository) scanVideoRow(rows pgx.Rows) (*entity.Video, error) {
	var v entity.Video
	var slug, contentRating, status, transcodeError string
	var duration int64

	err := rows.Scan(
		&v.ID, &v.Title, &slug, &v.Description, &v.Creator,
		&duration, &contentRating, &status,
		&v.StorageKey, &v.ContentType, &v.FileSize, &v.ThumbnailKey,
		&v.UploadedBy, &v.MinimumPlanLevel, &transcodeError, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning video row: %w", err)
	}

	v.Slug = valueobject.SlugFromTrusted(slug)
	v.Duration = valueobject.NewDuration(duration)
	v.ContentRating = contentRating
	v.Status = entity.VideoStatus(status)
	v.TranscodeError = transcodeError

	return &v, nil
}
