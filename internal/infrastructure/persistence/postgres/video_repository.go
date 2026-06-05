package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain"
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

const videoSelectCols = `
	id, title, slug, description, logline, creator, duration, content_rating,
	primary_language, status, visibility,
	storage_key, content_type, file_size, thumbnail_key, thumbnail_candidates,
	uploaded_by, minimum_plan_level, transcode_error,
	series_id, season_number, episode_number,
	scheduled_publish_at,
	monetization_types, ppv_price_cents, free_preview_seconds, ad_break_positions,
	geo_restriction_type, geo_restriction_regions, require_signin,
	submission_status, submitted_at, moderator_notes,
	is_featured, featured_description, featured_at,
	created_at, updated_at`

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

	monetizationJSON, _ := json.Marshal(video.MonetizationTypes)
	adBreakJSON, _ := json.Marshal(video.AdBreakPositions)
	geoRegionsJSON, _ := json.Marshal(video.GeoRestrictionRegions)
	thumbnailCandidatesJSON, _ := json.Marshal(video.ThumbnailCandidates)

	_, err := db.Exec(ctx, `
		INSERT INTO videos (
			id, title, slug, description, logline, creator, duration, content_rating,
			primary_language, status, visibility,
			storage_key, content_type, file_size, thumbnail_key, thumbnail_candidates,
			uploaded_by, minimum_plan_level, transcode_error,
			monetization_types, ppv_price_cents, free_preview_seconds, ad_break_positions,
			geo_restriction_type, geo_restriction_regions, require_signin,
			submission_status, moderator_notes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19,
			$20, $21, $22, $23,
			$24, $25, $26,
			$27, $28,
			$29, $30
		)`,
		video.ID, video.Title, video.Slug.String(), video.Description, video.Logline,
		video.Creator, video.Duration.Seconds, video.ContentRating,
		video.PrimaryLanguage, string(video.Status), string(video.Visibility),
		video.StorageKey, video.ContentType, video.FileSize, video.ThumbnailKey, thumbnailCandidatesJSON,
		video.UploadedBy, video.MinimumPlanLevel, video.TranscodeError,
		monetizationJSON, video.PPVPriceCents, video.FreePreviewSeconds, adBreakJSON,
		string(video.GeoRestrictionType), geoRegionsJSON, video.RequireSignin,
		string(video.SubmissionStatus), video.ModeratorNotes,
		video.CreatedAt, video.UpdatedAt,
	)
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
	return r.scanVideo(ctx, db, `SELECT `+videoSelectCols+` FROM videos v WHERE id = $1`, id)
}

func (r *VideoRepository) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Video, error) {
	db := connFromCtx(ctx, r.pool)
	return r.scanVideo(ctx, db, `SELECT `+videoSelectCols+` FROM videos v WHERE slug = $1`, slug.String())
}

func (r *VideoRepository) Update(ctx context.Context, video *entity.Video) error {
	db := connFromCtx(ctx, r.pool)

	monetizationJSON, _ := json.Marshal(video.MonetizationTypes)
	adBreakJSON, _ := json.Marshal(video.AdBreakPositions)
	geoRegionsJSON, _ := json.Marshal(video.GeoRestrictionRegions)
	thumbnailCandidatesJSON, _ := json.Marshal(video.ThumbnailCandidates)

	_, err := db.Exec(ctx, `
		UPDATE videos SET
			title = $2, slug = $3, description = $4, logline = $5, creator = $6,
			duration = $7, content_rating = $8, primary_language = $9,
			status = $10, visibility = $11,
			storage_key = $12, content_type = $13, file_size = $14, thumbnail_key = $15,
			thumbnail_candidates = $16,
			minimum_plan_level = $17, transcode_error = $18,
			scheduled_publish_at = $19,
			monetization_types = $20, ppv_price_cents = $21, free_preview_seconds = $22,
			ad_break_positions = $23,
			geo_restriction_type = $24, geo_restriction_regions = $25, require_signin = $26,
			submission_status = $27, submitted_at = $28, moderator_notes = $29,
			updated_at = $30
		WHERE id = $1`,
		video.ID, video.Title, video.Slug.String(), video.Description, video.Logline,
		video.Creator, video.Duration.Seconds, video.ContentRating, video.PrimaryLanguage,
		string(video.Status), string(video.Visibility),
		video.StorageKey, video.ContentType, video.FileSize, video.ThumbnailKey,
		thumbnailCandidatesJSON,
		video.MinimumPlanLevel, video.TranscodeError,
		video.ScheduledPublishAt,
		monetizationJSON, video.PPVPriceCents, video.FreePreviewSeconds,
		adBreakJSON,
		string(video.GeoRestrictionType), geoRegionsJSON, video.RequireSignin,
		string(video.SubmissionStatus), video.SubmittedAt, video.ModeratorNotes,
		video.UpdatedAt,
	)
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
	return r.listWithBase(ctx, fs, "v.status = 'published' AND v.series_id IS NULL")
}

func (r *VideoRepository) ListByUploader(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	db := connFromCtx(ctx, r.pool)

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
	dataQuery := fmt.Sprintf(`SELECT `+videoSelectCols+` FROM videos v %s %s %s`,
		sql.WhereClause, sql.OrderClause, sql.LimitClause)

	rows, err := db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing videos by uploader: %w", err)
	}
	defer rows.Close()

	videos, err := r.collectVideoRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

func (r *VideoRepository) listWithBase(ctx context.Context, fs *filter.FilterSet, baseCondition string) ([]entity.Video, int64, error) {
	db := connFromCtx(ctx, r.pool)

	countWhere, countArgs, _ := filter.CountSQL(fs, baseCondition, 1)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM videos v %s`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting videos: %w", err)
	}

	sql := filter.ToSQL(fs, baseCondition, 1)
	dataQuery := fmt.Sprintf(`SELECT `+videoSelectCols+` FROM videos v %s %s %s`,
		sql.WhereClause, sql.OrderClause, sql.LimitClause)

	rows, err := db.Query(ctx, dataQuery, sql.Args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing videos: %w", err)
	}
	defer rows.Close()

	videos, err := r.collectVideoRows(ctx, rows)
	if err != nil {
		return nil, 0, err
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
		_, err := db.Exec(ctx, `INSERT INTO video_genres (video_id, genre_id) VALUES ($1, $2)`, videoID, genreID)
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

func (r *VideoRepository) SetSeriesEpisode(ctx context.Context, videoID, seriesID uuid.UUID, season, episode int) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE videos SET series_id = $2, season_number = $3, episode_number = $4, updated_at = NOW()
		WHERE id = $1
	`, videoID, seriesID, season, episode)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEpisodeAlreadyExists
		}
		return fmt.Errorf("setting series episode: %w", err)
	}
	return nil
}

func (r *VideoRepository) RemoveFromSeries(ctx context.Context, videoID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE videos SET series_id = NULL, season_number = NULL, episode_number = NULL, updated_at = NOW()
		WHERE id = $1
	`, videoID)
	if err != nil {
		return fmt.Errorf("removing video from series: %w", err)
	}
	return nil
}

func (r *VideoRepository) ListBySeries(ctx context.Context, seriesID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	db := connFromCtx(ctx, r.pool)

	baseCondition := "v.series_id = $1"
	baseArgs := []any{seriesID}

	countWhere, countFilterArgs, _ := filter.CountSQL(fs, baseCondition, 2)
	countArgs := append(append([]any{}, baseArgs...), countFilterArgs...)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM videos v %s`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting videos by series: %w", err)
	}

	sql := filter.ToSQL(fs, baseCondition, 2)
	dataArgs := append(append([]any{}, baseArgs...), sql.Args...)
	orderClause := "ORDER BY v.season_number ASC NULLS LAST, v.episode_number ASC NULLS LAST"
	dataQuery := fmt.Sprintf(`SELECT `+videoSelectCols+` FROM videos v %s %s %s`,
		sql.WhereClause, orderClause, sql.LimitClause)

	rows, err := db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing videos by series: %w", err)
	}
	defer rows.Close()

	videos, err := r.collectVideoRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// collectVideoRows scans all rows and loads genre IDs for each video.
func (r *VideoRepository) collectVideoRows(ctx context.Context, rows pgx.Rows) ([]entity.Video, error) {
	var videos []entity.Video
	for rows.Next() {
		v, err := r.scanVideoRow(rows)
		if err != nil {
			return nil, err
		}
		videos = append(videos, *v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range videos {
		genreIDs, err := r.GetGenreIDs(ctx, videos[i].ID)
		if err != nil {
			return nil, err
		}
		videos[i].GenreIDs = genreIDs
	}

	return videos, nil
}

func (r *VideoRepository) scanVideo(ctx context.Context, db DBTX, query string, args ...any) (*entity.Video, error) {
	row := db.QueryRow(ctx, query, args...)

	var (
		v                                                   entity.Video
		slug, status, visibility, geoType, submissionStatus string
		duration                                            int64
		monetizationJSON                                    json.RawMessage
		adBreakJSON                                         json.RawMessage
		geoRegionsJSON                                      json.RawMessage
		thumbnailCandidatesJSON                             json.RawMessage
	)

	err := row.Scan(
		&v.ID, &v.Title, &slug, &v.Description, &v.Logline, &v.Creator,
		&duration, &v.ContentRating,
		&v.PrimaryLanguage, &status, &visibility,
		&v.StorageKey, &v.ContentType, &v.FileSize, &v.ThumbnailKey, &thumbnailCandidatesJSON,
		&v.UploadedBy, &v.MinimumPlanLevel, &v.TranscodeError,
		&v.SeriesID, &v.SeasonNumber, &v.EpisodeNumber,
		&v.ScheduledPublishAt,
		&monetizationJSON, &v.PPVPriceCents, &v.FreePreviewSeconds, &adBreakJSON,
		&geoType, &geoRegionsJSON, &v.RequireSignin,
		&submissionStatus, &v.SubmittedAt, &v.ModeratorNotes,
		&v.IsFeatured, &v.FeaturedDescription, &v.FeaturedAt,
		&v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning video: %w", err)
	}

	v.Slug = valueobject.SlugFromTrusted(slug)
	v.Duration = valueobject.NewDuration(duration)
	v.Status = entity.VideoStatus(status)
	v.Visibility = entity.VideoVisibility(visibility)
	v.GeoRestrictionType = entity.GeoRestrictionType(geoType)
	v.SubmissionStatus = entity.SubmissionStatus(submissionStatus)
	unmarshalJSONOrEmpty(monetizationJSON, &v.MonetizationTypes)
	unmarshalJSONOrEmpty(adBreakJSON, &v.AdBreakPositions)
	unmarshalJSONOrEmpty(geoRegionsJSON, &v.GeoRestrictionRegions)
	unmarshalJSONOrEmpty(thumbnailCandidatesJSON, &v.ThumbnailCandidates)

	genreIDs, err := r.GetGenreIDs(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	v.GenreIDs = genreIDs

	return &v, nil
}

func (r *VideoRepository) scanVideoRow(rows pgx.Rows) (*entity.Video, error) {
	var (
		v                                                   entity.Video
		slug, status, visibility, geoType, submissionStatus string
		duration                                            int64
		monetizationJSON                                    json.RawMessage
		adBreakJSON                                         json.RawMessage
		geoRegionsJSON                                      json.RawMessage
		thumbnailCandidatesJSON                             json.RawMessage
	)

	err := rows.Scan(
		&v.ID, &v.Title, &slug, &v.Description, &v.Logline, &v.Creator,
		&duration, &v.ContentRating,
		&v.PrimaryLanguage, &status, &visibility,
		&v.StorageKey, &v.ContentType, &v.FileSize, &v.ThumbnailKey, &thumbnailCandidatesJSON,
		&v.UploadedBy, &v.MinimumPlanLevel, &v.TranscodeError,
		&v.SeriesID, &v.SeasonNumber, &v.EpisodeNumber,
		&v.ScheduledPublishAt,
		&monetizationJSON, &v.PPVPriceCents, &v.FreePreviewSeconds, &adBreakJSON,
		&geoType, &geoRegionsJSON, &v.RequireSignin,
		&submissionStatus, &v.SubmittedAt, &v.ModeratorNotes,
		&v.IsFeatured, &v.FeaturedDescription, &v.FeaturedAt,
		&v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning video row: %w", err)
	}

	v.Slug = valueobject.SlugFromTrusted(slug)
	v.Duration = valueobject.NewDuration(duration)
	v.Status = entity.VideoStatus(status)
	v.Visibility = entity.VideoVisibility(visibility)
	v.GeoRestrictionType = entity.GeoRestrictionType(geoType)
	v.SubmissionStatus = entity.SubmissionStatus(submissionStatus)
	unmarshalJSONOrEmpty(monetizationJSON, &v.MonetizationTypes)
	unmarshalJSONOrEmpty(adBreakJSON, &v.AdBreakPositions)
	unmarshalJSONOrEmpty(geoRegionsJSON, &v.GeoRestrictionRegions)
	unmarshalJSONOrEmpty(thumbnailCandidatesJSON, &v.ThumbnailCandidates)

	return &v, nil
}

func (r *VideoRepository) ListFeatured(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	return r.listWithBase(ctx, fs, "v.is_featured = TRUE AND v.status = 'published'")
}

func (r *VideoRepository) SetFeatured(ctx context.Context, videoID uuid.UUID, description string) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE videos
		SET is_featured = TRUE, featured_description = $2, featured_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, videoID, description)
	if err != nil {
		return fmt.Errorf("setting video as featured: %w", err)
	}
	return nil
}

func (r *VideoRepository) UnsetFeatured(ctx context.Context, videoID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE videos
		SET is_featured = FALSE, featured_description = NULL, featured_at = NULL, updated_at = NOW()
		WHERE id = $1
	`, videoID)
	if err != nil {
		return fmt.Errorf("unsetting video as featured: %w", err)
	}
	return nil
}

func unmarshalJSONOrEmpty[T any](data json.RawMessage, dest *T) {
	if len(data) > 0 {
		_ = json.Unmarshal(data, dest)
	}
}
