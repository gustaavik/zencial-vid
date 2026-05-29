package series

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// AddEpisodeInput holds the data needed to add a video as an episode.
type AddEpisodeInput struct {
	SeriesID      uuid.UUID
	VideoID       uuid.UUID
	SeasonNumber  int
	EpisodeNumber int
	CallerID      uuid.UUID
	CallerRoles   []entity.UserRole
}

// RemoveEpisodeInput holds the data needed to remove an episode from a series.
type RemoveEpisodeInput struct {
	SeriesID    uuid.UUID
	VideoID     uuid.UUID
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
}

// AddEpisode links a video to a series as an ordered episode.
func (s *Service) AddEpisode(ctx context.Context, input *AddEpisodeInput) (*entity.Video, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, input.SeriesID)
	if err != nil {
		s.log.Error("getting series for add episode", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) && series.UploadedBy != input.CallerID {
		return nil, apperror.Forbidden(apperror.CodeSeriesOwnershipRequired, "you do not own this series", domain.ErrSeriesOwnershipRequired)
	}

	video, err := s.videoRepo.GetByID(ctx, input.VideoID)
	if err != nil {
		s.log.Error("getting video for add episode", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if video.SeriesID != nil && *video.SeriesID != input.SeriesID {
		return nil, apperror.Conflict(apperror.CodeEpisodeAlreadyExists, "video already belongs to another series", domain.ErrEpisodeAlreadyExists)
	}

	if err := s.videoRepo.SetSeriesEpisode(ctx, input.VideoID, input.SeriesID, input.SeasonNumber, input.EpisodeNumber); err != nil {
		s.log.Error("setting series episode", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to add episode", err)
	}

	video.SeriesID = &input.SeriesID
	video.SeasonNumber = &input.SeasonNumber
	video.EpisodeNumber = &input.EpisodeNumber

	if err := s.dispatcher.Dispatch(event.SeriesUpdated{
		SeriesID:  series.ID,
		Field:     "episodes",
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching series updated event", "error", err)
	}

	return video, nil
}

// RemoveEpisode unlinks a video from a series.
func (s *Service) RemoveEpisode(ctx context.Context, input *RemoveEpisodeInput) *apperror.AppError {
	series, err := s.seriesRepo.GetByID(ctx, input.SeriesID)
	if err != nil {
		s.log.Error("getting series for remove episode", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) && series.UploadedBy != input.CallerID {
		return apperror.Forbidden(apperror.CodeSeriesOwnershipRequired, "you do not own this series", domain.ErrSeriesOwnershipRequired)
	}

	video, err := s.videoRepo.GetByID(ctx, input.VideoID)
	if err != nil {
		s.log.Error("getting video for remove episode", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if video.SeriesID == nil || *video.SeriesID != input.SeriesID {
		return apperror.BadRequest(apperror.CodeBadRequest, "video does not belong to this series", nil)
	}

	if err := s.videoRepo.RemoveFromSeries(ctx, input.VideoID); err != nil {
		s.log.Error("removing video from series", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to remove episode", err)
	}

	if err := s.dispatcher.Dispatch(event.SeriesUpdated{
		SeriesID:  series.ID,
		Field:     "episodes",
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching series updated event", "error", err)
	}

	return nil
}

// ListEpisodes returns all episodes of a series ordered by season/episode number.
func (s *Service) ListEpisodes(ctx context.Context, seriesID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		s.log.Error("getting series for list episodes", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, 0, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	videos, total, err := s.videoRepo.ListBySeries(ctx, seriesID, fs)
	if err != nil {
		s.log.Error("listing series episodes", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list episodes", err)
	}
	return videos, total, nil
}
