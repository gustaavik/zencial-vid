package video

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateInput holds the data needed to update a video's metadata.
type UpdateInput struct {
	ID               uuid.UUID
	Title            *string
	Description      *string
	Creator          *string
	ContentRating    *string
	GenreIDs         []uuid.UUID
	MinimumPlanLevel *int
}

// Update updates a video's metadata.
func (s *Service) Update(ctx context.Context, input *UpdateInput) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, input.ID)
	if err != nil {
		s.log.Error("getting video for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if input.Title != nil {
		video.Title = *input.Title
		slug, err := valueobject.NewSlug(*input.Title)
		if err != nil {
			return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid title for slug generation", err)
		}
		video.Slug = slug.WithRandomID()
	}

	if input.Description != nil {
		video.Description = *input.Description
	}

	if input.Creator != nil {
		video.Creator = *input.Creator
	}

	if input.ContentRating != nil {
		video.ContentRating = *input.ContentRating
	}

	if input.MinimumPlanLevel != nil {
		video.MinimumPlanLevel = input.MinimumPlanLevel
	}

	video.UpdatedAt = time.Now().UTC()
	if err := s.videoRepo.Update(ctx, video); err != nil {
		s.log.Error("updating video", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update video", err)
	}

	if input.GenreIDs != nil {
		if err := s.videoRepo.SetGenres(ctx, video.ID, input.GenreIDs); err != nil {
			s.log.Error("setting video genres", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to update video genres", err)
		}
		video.GenreIDs = input.GenreIDs
	}

	if err := s.dispatcher.Dispatch(event.VideoUpdated{
		VideoID:   video.ID,
		ActorID:   actor.FromContext(ctx),
		Field:     "metadata",
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching video updated event", "error", err)
	}

	return video, nil
}
