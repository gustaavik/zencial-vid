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

// MusicCueRepository implements repository.MusicCueRepository using PostgreSQL.
type MusicCueRepository struct {
	pool *pgxpool.Pool
}

// NewMusicCueRepository creates a new MusicCueRepository.
func NewMusicCueRepository(pool *pgxpool.Pool) *MusicCueRepository {
	return &MusicCueRepository{pool: pool}
}

func (r *MusicCueRepository) Create(ctx context.Context, cue *entity.MusicCue) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO music_cues (
			id, video_id, timecode_seconds, title, composer_artist,
			use_type, rights_status, clearance_document_key,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, cue.ID, cue.VideoID, cue.TimecodeSeconds, cue.Title, cue.ComposerArtist,
		string(cue.UseType), string(cue.RightsStatus), cue.ClearanceDocumentKey,
		cue.CreatedAt, cue.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating music cue: %w", err)
	}
	return nil
}

func (r *MusicCueRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.MusicCue, error) {
	db := connFromCtx(ctx, r.pool)
	var cue entity.MusicCue
	var useType, rightsStatus string
	err := db.QueryRow(ctx, `
		SELECT id, video_id, timecode_seconds, title, composer_artist,
		       use_type, rights_status, clearance_document_key, created_at, updated_at
		FROM music_cues WHERE id = $1
	`, id).Scan(
		&cue.ID, &cue.VideoID, &cue.TimecodeSeconds, &cue.Title, &cue.ComposerArtist,
		&useType, &rightsStatus, &cue.ClearanceDocumentKey, &cue.CreatedAt, &cue.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting music cue: %w", err)
	}
	cue.UseType = entity.MusicUseType(useType)
	cue.RightsStatus = entity.MusicRightsStatus(rightsStatus)
	return &cue, nil
}

func (r *MusicCueRepository) ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.MusicCue, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, video_id, timecode_seconds, title, composer_artist,
		       use_type, rights_status, clearance_document_key, created_at, updated_at
		FROM music_cues WHERE video_id = $1
		ORDER BY timecode_seconds ASC
	`, videoID)
	if err != nil {
		return nil, fmt.Errorf("listing music cues: %w", err)
	}
	defer rows.Close()

	var cues []entity.MusicCue
	for rows.Next() {
		var cue entity.MusicCue
		var useType, rightsStatus string
		if err := rows.Scan(
			&cue.ID, &cue.VideoID, &cue.TimecodeSeconds, &cue.Title, &cue.ComposerArtist,
			&useType, &rightsStatus, &cue.ClearanceDocumentKey, &cue.CreatedAt, &cue.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning music cue: %w", err)
		}
		cue.UseType = entity.MusicUseType(useType)
		cue.RightsStatus = entity.MusicRightsStatus(rightsStatus)
		cues = append(cues, cue)
	}
	return cues, rows.Err()
}

func (r *MusicCueRepository) Update(ctx context.Context, cue *entity.MusicCue) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE music_cues SET
			timecode_seconds = $2, title = $3, composer_artist = $4,
			use_type = $5, rights_status = $6, clearance_document_key = $7,
			updated_at = $8
		WHERE id = $1
	`, cue.ID, cue.TimecodeSeconds, cue.Title, cue.ComposerArtist,
		string(cue.UseType), string(cue.RightsStatus), cue.ClearanceDocumentKey,
		cue.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating music cue: %w", err)
	}
	return nil
}

func (r *MusicCueRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM music_cues WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting music cue: %w", err)
	}
	return nil
}

func (r *MusicCueRepository) HasBlockingCues(ctx context.Context, videoID uuid.UUID) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var has bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM music_cues
			WHERE video_id = $1 AND rights_status = 'pending_clearance'
		)
	`, videoID).Scan(&has)
	if err != nil {
		return false, fmt.Errorf("checking blocking music cues: %w", err)
	}
	return has, nil
}
