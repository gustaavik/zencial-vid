package season

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CreateSeasonInput holds inputs for creating a season.
type CreateSeasonInput struct {
	SeriesID        uuid.UUID
	UploaderID      uuid.UUID
	SeasonNumber    int
	SeasonTag       string
	PlannedEpisodes int
	AvgRuntimeSecs  int
	ReleaseCadence  entity.ReleaseCadence
	PremiereDate    *time.Time
	CadenceDay      *int
	Timezone        string
}

// CreateSeason creates a new season for a series.
func (s *Service) CreateSeason(ctx context.Context, input *CreateSeasonInput) (*entity.Season, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, input.SeriesID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", nil)
	}
	if series.UploadedBy != input.UploaderID {
		return nil, apperror.Forbidden(apperror.CodeSeriesOwnershipRequired, "you do not own this series", nil)
	}

	season := entity.NewSeason(input.SeriesID, input.SeasonNumber)
	season.SeasonTag = input.SeasonTag
	season.PlannedEpisodes = input.PlannedEpisodes
	season.AvgRuntimeSecs = input.AvgRuntimeSecs
	season.ReleaseCadence = input.ReleaseCadence
	season.PremiereDate = input.PremiereDate
	season.CadenceDay = input.CadenceDay
	if input.Timezone != "" {
		season.Timezone = input.Timezone
	}

	if err := s.seasonRepo.Create(ctx, season); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create season", err)
	}
	return season, nil
}
