package musiccue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CreateCueInput holds inputs for adding a music cue.
type CreateCueInput struct {
	VideoID         uuid.UUID
	UploaderID      uuid.UUID
	TimecodeSeconds int
	Title           string
	ComposerArtist  string
	UseType         entity.MusicUseType
	RightsStatus    entity.MusicRightsStatus
}

// CreateCue adds a music cue to a video.
func (s *Service) CreateCue(ctx context.Context, input *CreateCueInput) (*entity.MusicCue, *apperror.AppError) {
	if err := s.assertOwnership(ctx, input.VideoID, input.UploaderID); err != nil {
		return nil, err
	}

	cue := entity.NewMusicCue(input.VideoID, input.TimecodeSeconds, input.Title, input.ComposerArtist, input.UseType)
	cue.RightsStatus = input.RightsStatus

	if err := s.cueRepo.Create(ctx, cue); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create music cue", err)
	}
	return cue, nil
}

// UpdateCueInput holds inputs for updating a music cue.
type UpdateCueInput struct {
	CueID           uuid.UUID
	UploaderID      uuid.UUID
	TimecodeSeconds *int
	Title           *string
	ComposerArtist  *string
	UseType         *entity.MusicUseType
	RightsStatus    *entity.MusicRightsStatus
}

// UpdateCue updates a music cue's metadata.
func (s *Service) UpdateCue(ctx context.Context, input *UpdateCueInput) (*entity.MusicCue, *apperror.AppError) {
	cue, err := s.cueRepo.GetByID(ctx, input.CueID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch music cue", err)
	}
	if cue == nil {
		return nil, apperror.NotFound(apperror.CodeMusicCueNotFound, "music cue not found", nil)
	}

	if appErr := s.assertOwnership(ctx, cue.VideoID, input.UploaderID); appErr != nil {
		return nil, appErr
	}

	if input.TimecodeSeconds != nil {
		cue.TimecodeSeconds = *input.TimecodeSeconds
	}
	if input.Title != nil {
		cue.Title = *input.Title
	}
	if input.ComposerArtist != nil {
		cue.ComposerArtist = *input.ComposerArtist
	}
	if input.UseType != nil {
		cue.UseType = *input.UseType
	}
	if input.RightsStatus != nil {
		cue.RightsStatus = *input.RightsStatus
	}
	cue.UpdatedAt = time.Now().UTC()

	if err := s.cueRepo.Update(ctx, cue); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update music cue", err)
	}
	return cue, nil
}

// ListCues returns all music cues for a video ordered by timecode.
func (s *Service) ListCues(ctx context.Context, videoID uuid.UUID) ([]entity.MusicCue, *apperror.AppError) {
	cues, err := s.cueRepo.ListByVideo(ctx, videoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list music cues", err)
	}
	return cues, nil
}

// DeleteCue removes a music cue.
func (s *Service) DeleteCue(ctx context.Context, cueID, uploaderID uuid.UUID) *apperror.AppError {
	cue, err := s.cueRepo.GetByID(ctx, cueID)
	if err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch music cue", err)
	}
	if cue == nil {
		return apperror.NotFound(apperror.CodeMusicCueNotFound, "music cue not found", nil)
	}

	if appErr := s.assertOwnership(ctx, cue.VideoID, uploaderID); appErr != nil {
		return appErr
	}

	if err := s.cueRepo.DeleteByID(ctx, cueID); err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to delete music cue", err)
	}
	return nil
}

// InitiateClearanceUploadOutput contains the signed upload URL for a clearance document.
type InitiateClearanceUploadOutput struct {
	UploadURL string
	ObjectKey string
	ExpiresAt time.Time
}

// InitiateClearanceUpload returns a signed PUT URL for uploading a clearance document.
func (s *Service) InitiateClearanceUpload(ctx context.Context, cueID, uploaderID uuid.UUID) (*InitiateClearanceUploadOutput, *apperror.AppError) {
	cue, err := s.cueRepo.GetByID(ctx, cueID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch music cue", err)
	}
	if cue == nil {
		return nil, apperror.NotFound(apperror.CodeMusicCueNotFound, "music cue not found", nil)
	}

	if appErr := s.assertOwnership(ctx, cue.VideoID, uploaderID); appErr != nil {
		return nil, appErr
	}

	objectKey := fmt.Sprintf("clearances/%s/%s.pdf", cue.VideoID, cue.ID)
	expiry := 30 * time.Minute
	uploadURL, signErr := s.storage.PresignedPutURL(ctx, objectKey, "application/pdf", expiry)
	if signErr != nil {
		return nil, apperror.Internal(apperror.CodeStorageError, "failed to sign clearance upload URL", signErr)
	}

	return &InitiateClearanceUploadOutput{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		ExpiresAt: time.Now().UTC().Add(expiry),
	}, nil
}

// CompleteClearanceUpload records the clearance document key and marks the cue as cleared.
func (s *Service) CompleteClearanceUpload(ctx context.Context, cueID uuid.UUID, objectKey string, uploaderID uuid.UUID) (*entity.MusicCue, *apperror.AppError) {
	cue, err := s.cueRepo.GetByID(ctx, cueID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch music cue", err)
	}
	if cue == nil {
		return nil, apperror.NotFound(apperror.CodeMusicCueNotFound, "music cue not found", nil)
	}

	if appErr := s.assertOwnership(ctx, cue.VideoID, uploaderID); appErr != nil {
		return nil, appErr
	}

	cue.AttachClearanceDocument(objectKey)
	if err := s.cueRepo.Update(ctx, cue); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update music cue", err)
	}
	return cue, nil
}

func (s *Service) assertOwnership(ctx context.Context, videoID, uploaderID uuid.UUID) *apperror.AppError {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", nil)
	}
	if video.UploadedBy != uploaderID {
		return apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil)
	}
	return nil
}
