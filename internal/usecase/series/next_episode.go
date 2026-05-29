package series

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// GetNextEpisode returns the next episode to watch for a user in a series.
// If the user has no progress, the first episode is returned.
// Returns nil, nil when the series has no episodes.
func (s *Service) GetNextEpisode(ctx context.Context, userID, seriesID uuid.UUID) (*entity.Video, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		s.log.Error("getting series for next episode", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	// Load all episodes ordered by season/episode ascending.
	fs := &filter.FilterSet{Pagination: valueobject.NewPagination(1, 500)}
	episodes, _, err := s.videoRepo.ListBySeries(ctx, seriesID, fs)
	if err != nil {
		s.log.Error("listing series episodes for next episode", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get episodes", err)
	}
	if len(episodes) == 0 {
		return nil, apperror.NotFound(apperror.CodeNotFound, "series has no episodes", nil)
	}

	progress, err := s.seriesWpRepo.Get(ctx, userID, seriesID)
	if err != nil {
		s.log.Error("getting series watch progress", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get watch progress", err)
	}

	// No progress — return the first episode.
	if progress == nil {
		return &episodes[0], nil
	}

	// Find the episode after the last-watched one.
	for i := range episodes {
		if episodes[i].ID == progress.LastEpisodeID && i+1 < len(episodes) {
			return &episodes[i+1], nil
		}
	}

	// All episodes watched — return not found.
	return nil, apperror.NotFound(apperror.CodeNotFound, "no next episode", nil)
}
