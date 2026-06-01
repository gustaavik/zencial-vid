package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// VideoCastRepository implements repository.VideoCastRepository using PostgreSQL.
type VideoCastRepository struct {
	pool *pgxpool.Pool
}

// NewVideoCastRepository creates a new VideoCastRepository.
func NewVideoCastRepository(pool *pgxpool.Pool) *VideoCastRepository {
	return &VideoCastRepository{pool: pool}
}

func (r *VideoCastRepository) Create(ctx context.Context, vc *entity.VideoCast) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO video_cast (
			id, video_id, cast_id, role, department,
			invite_status, invited_email, sort_order,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, vc.ID, vc.VideoID, vc.CastID, vc.Role, string(vc.Department),
		string(vc.InviteStatus), vc.InvitedEmail, vc.SortOrder,
		vc.CreatedAt, vc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating video cast: %w", err)
	}
	return nil
}

func (r *VideoCastRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.VideoCast, error) {
	db := connFromCtx(ctx, r.pool)
	return scanVideoCast(db.QueryRow(ctx, `
		SELECT
			vc.id, vc.video_id, vc.cast_id, vc.role, vc.department,
			vc.invite_status, vc.invited_email, vc.sort_order,
			vc.created_at, vc.updated_at,
			c.id, c.name, c.picture_key, c.created_at, c.updated_at
		FROM video_cast vc
		JOIN casts c ON c.id = vc.cast_id
		WHERE vc.id = $1
	`, id))
}

func (r *VideoCastRepository) GetByVideoAndCastAndRole(ctx context.Context, videoID, castID uuid.UUID, role string) (*entity.VideoCast, error) {
	db := connFromCtx(ctx, r.pool)
	return scanVideoCast(db.QueryRow(ctx, `
		SELECT
			vc.id, vc.video_id, vc.cast_id, vc.role, vc.department,
			vc.invite_status, vc.invited_email, vc.sort_order,
			vc.created_at, vc.updated_at,
			c.id, c.name, c.picture_key, c.created_at, c.updated_at
		FROM video_cast vc
		JOIN casts c ON c.id = vc.cast_id
		WHERE vc.video_id = $1 AND vc.cast_id = $2 AND vc.role = $3
	`, videoID, castID, role))
}

func (r *VideoCastRepository) ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.VideoCast, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT
			vc.id, vc.video_id, vc.cast_id, vc.role, vc.department,
			vc.invite_status, vc.invited_email, vc.sort_order,
			vc.created_at, vc.updated_at,
			c.id, c.name, c.picture_key, c.created_at, c.updated_at
		FROM video_cast vc
		JOIN casts c ON c.id = vc.cast_id
		WHERE vc.video_id = $1
		ORDER BY vc.sort_order ASC, vc.created_at ASC
	`, videoID)
	if err != nil {
		return nil, fmt.Errorf("listing video cast: %w", err)
	}
	defer rows.Close()

	var results []entity.VideoCast
	for rows.Next() {
		vc, err := scanVideoCastRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *vc)
	}
	return results, rows.Err()
}

func (r *VideoCastRepository) Update(ctx context.Context, vc *entity.VideoCast) error {
	db := connFromCtx(ctx, r.pool)
	vc.UpdatedAt = time.Now().UTC()
	_, err := db.Exec(ctx, `
		UPDATE video_cast SET
			role = $2, department = $3,
			invite_status = $4, sort_order = $5,
			updated_at = $6
		WHERE id = $1
	`, vc.ID, vc.Role, string(vc.Department),
		string(vc.InviteStatus), vc.SortOrder, vc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating video cast: %w", err)
	}
	return nil
}

func (r *VideoCastRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM video_cast WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting video cast: %w", err)
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanVideoCast(row pgx.Row) (*entity.VideoCast, error) {
	return scanVideoCastRow(row)
}

func scanVideoCastRow(row scannable) (*entity.VideoCast, error) {
	vc := &entity.VideoCast{Cast: &entity.Cast{}}
	var department, inviteStatus string
	var pictureKey *string
	err := row.Scan(
		&vc.ID, &vc.VideoID, &vc.CastID, &vc.Role, &department,
		&inviteStatus, &vc.InvitedEmail, &vc.SortOrder,
		&vc.CreatedAt, &vc.UpdatedAt,
		&vc.Cast.ID, &vc.Cast.Name, &pictureKey, &vc.Cast.CreatedAt, &vc.Cast.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning video cast: %w", err)
	}
	vc.Department = entity.CastDepartment(department)
	vc.InviteStatus = entity.CastInviteStatus(inviteStatus)
	if pictureKey != nil {
		vc.Cast.PictureKey = *pictureKey
	}
	return vc, nil
}

func scanVideoCastWithVideo(row scannable) (*entity.VideoCast, error) {
	vc := &entity.VideoCast{Video: &entity.Video{}}
	var slug, status, visibility, geoType, submissionStatus, department, inviteStatus string
	var duration int64
	err := row.Scan(
		&vc.ID, &vc.VideoID, &vc.CastID, &vc.Role, &department,
		&inviteStatus, &vc.InvitedEmail, &vc.SortOrder,
		&vc.CreatedAt, &vc.UpdatedAt,
		&vc.Video.ID, &vc.Video.Title, &slug, &vc.Video.Description,
		&vc.Video.Logline, &vc.Video.Creator, &duration,
		&vc.Video.ContentRating, &status, &visibility,
		&vc.Video.StorageKey, &vc.Video.ContentType, &vc.Video.FileSize,
		&vc.Video.ThumbnailKey, &vc.Video.UploadedBy,
		&vc.Video.MinimumPlanLevel, &vc.Video.TranscodeError,
		&vc.Video.CreatedAt, &vc.Video.UpdatedAt,
		&vc.Video.SeriesID, &vc.Video.SeasonNumber, &vc.Video.EpisodeNumber,
		&geoType, &submissionStatus,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning video cast with video: %w", err)
	}
	vc.Department = entity.CastDepartment(department)
	vc.InviteStatus = entity.CastInviteStatus(inviteStatus)
	vc.Video.Slug = valueobject.SlugFromTrusted(slug)
	vc.Video.Duration = valueobject.NewDuration(duration)
	vc.Video.Status = entity.VideoStatus(status)
	vc.Video.Visibility = entity.VideoVisibility(visibility)
	vc.Video.GeoRestrictionType = entity.GeoRestrictionType(geoType)
	vc.Video.SubmissionStatus = entity.SubmissionStatus(submissionStatus)
	return vc, nil
}

func (r *VideoCastRepository) ListByCast(ctx context.Context, castID uuid.UUID, offset, limit int) ([]entity.VideoCast, int, error) {
	db := connFromCtx(ctx, r.pool)

	var total int
	if err := db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM video_cast vc
		JOIN videos v ON v.id = vc.video_id
		WHERE vc.cast_id = $1 AND v.status = 'published'
	`, castID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting videos by cast: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT
			vc.id, vc.video_id, vc.cast_id, vc.role, vc.department,
			vc.invite_status, vc.invited_email, vc.sort_order,
			vc.created_at, vc.updated_at,
			v.id, v.title, v.slug, v.description, v.logline, v.creator, v.duration,
			v.content_rating, v.status, v.visibility,
			v.storage_key, v.content_type, v.file_size, v.thumbnail_key, v.uploaded_by,
			v.minimum_plan_level, v.transcode_error, v.created_at, v.updated_at,
			v.series_id, v.season_number, v.episode_number,
			v.geo_restriction_type, v.submission_status
		FROM video_cast vc
		JOIN videos v ON v.id = vc.video_id
		WHERE vc.cast_id = $1 AND v.status = 'published'
		ORDER BY v.created_at DESC
		LIMIT $2 OFFSET $3
	`, castID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing videos by cast: %w", err)
	}
	defer rows.Close()

	var results []entity.VideoCast
	for rows.Next() {
		vc, err := scanVideoCastWithVideo(rows)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, *vc)
	}
	return results, total, rows.Err()
}
