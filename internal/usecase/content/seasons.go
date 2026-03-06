package content

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetSeasons returns all seasons for a series identified by slug.
func (s *Service) GetSeasons(ctx context.Context, slug string) ([]entity.Season, *apperror.AppError) {
	content, appErr := s.GetBySlug(ctx, slug)
	if appErr != nil {
		return nil, appErr
	}
	seasons, err := s.contentRepo.GetSeasonsForContent(ctx, content.ID)
	if err != nil {
		s.log.Error("getting seasons", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get seasons", err)
	}
	return seasons, nil
}

// GetEpisodes returns all episodes for a given season of a series.
func (s *Service) GetEpisodes(ctx context.Context, slug string, seasonNumber int) ([]entity.Episode, *apperror.AppError) {
	content, appErr := s.GetBySlug(ctx, slug)
	if appErr != nil {
		return nil, appErr
	}
	seasons, err := s.contentRepo.GetSeasonsForContent(ctx, content.ID)
	if err != nil {
		s.log.Error("getting seasons for episodes", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get seasons", err)
	}
	for _, season := range seasons {
		if season.Number == seasonNumber {
			episodes, err := s.contentRepo.GetEpisodesForSeason(ctx, season.ID)
			if err != nil {
				s.log.Error("getting episodes", "error", err)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to get episodes", err)
			}
			return episodes, nil
		}
	}
	return nil, apperror.NotFound(apperror.CodeSeasonNotFound, "season not found", domain.ErrSeasonNotFound)
}
