package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// SubmitForReview locks a video and queues it for moderation.
func (s *Service) SubmitForReview(ctx context.Context, videoID, uploaderID uuid.UUID) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", nil)
	}
	if video.UploadedBy != uploaderID {
		return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil)
	}
	if !video.CanBeEdited() {
		return nil, apperror.Conflict(apperror.CodeVideoAlreadySubmitted, "video is already submitted or under review", nil)
	}
	if video.Status == entity.VideoStatusDraft {
		return nil, apperror.BadRequest(apperror.CodeVideoNotSubmittable, "video must be encoded before submitting", nil)
	}

	// Check for blocking music cues via the music cue repo.
	if s.musicCueRepo != nil {
		hasBlocking, err := s.musicCueRepo.HasBlockingCues(ctx, videoID)
		if err != nil {
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to check music cues", err)
		}
		if hasBlocking {
			return nil, apperror.Conflict(apperror.CodeMusicCueBlocksSubmission, "one or more music cues have pending clearance", nil)
		}
	}

	video.Submit()
	if err := s.videoRepo.Update(ctx, video); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to submit video", err)
	}
	return video, nil
}

// ApproveSubmission approves a video submission and triggers the publish flow.
func (s *Service) ApproveSubmission(ctx context.Context, videoID uuid.UUID) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", nil)
	}
	if video.SubmissionStatus != entity.SubmissionStatusSubmitted &&
		video.SubmissionStatus != entity.SubmissionStatusUnderReview {
		return nil, apperror.BadRequest(apperror.CodeModerationNotFound, "video is not pending approval", nil)
	}

	video.ApproveSubmission()
	video.Publish() // kick off transcode/publish pipeline
	if err := s.videoRepo.Update(ctx, video); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to approve video", err)
	}

	if err := s.cdn.TriggerTranscode(videoID.String()); err != nil {
		s.log.Error("failed to trigger transcode after approval", "video_id", videoID, "err", err)
	}

	return video, nil
}

// RejectSubmission rejects a video submission with moderator notes.
func (s *Service) RejectSubmission(ctx context.Context, videoID uuid.UUID, notes string) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", nil)
	}

	video.RejectSubmission(notes)
	if err := s.videoRepo.Update(ctx, video); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to reject video", err)
	}
	return video, nil
}
