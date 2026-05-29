package season

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateSeasonInput holds inputs for updating a season.
type UpdateSeasonInput struct {
	SeasonID        uuid.UUID
	UploaderID      uuid.UUID
	SeasonTag       *string
	PlannedEpisodes *int
	AvgRuntimeSecs  *int
	ReleaseCadence  *entity.ReleaseCadence
	PremiereDate    *time.Time
	CadenceDay      *int
	Timezone        *string
}

// UpdateSeason updates an existing season's metadata.
func (s *Service) UpdateSeason(ctx context.Context, input *UpdateSeasonInput) (*entity.Season, *apperror.AppError) {
	season, err := s.seasonRepo.GetByID(ctx, input.SeasonID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch season", err)
	}
	if season == nil {
		return nil, apperror.NotFound(apperror.CodeSeasonNotFound, "season not found", nil)
	}

	series, err := s.seriesRepo.GetByID(ctx, season.SeriesID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch series", err)
	}
	if series == nil || series.UploadedBy != input.UploaderID {
		return nil, apperror.Forbidden(apperror.CodeSeriesOwnershipRequired, "you do not own this series", nil)
	}

	if input.SeasonTag != nil {
		season.SeasonTag = *input.SeasonTag
	}
	if input.PlannedEpisodes != nil {
		season.PlannedEpisodes = *input.PlannedEpisodes
	}
	if input.AvgRuntimeSecs != nil {
		season.AvgRuntimeSecs = *input.AvgRuntimeSecs
	}
	if input.ReleaseCadence != nil {
		season.ReleaseCadence = *input.ReleaseCadence
	}
	if input.PremiereDate != nil {
		season.PremiereDate = input.PremiereDate
	}
	if input.CadenceDay != nil {
		season.CadenceDay = input.CadenceDay
	}
	if input.Timezone != nil {
		season.Timezone = *input.Timezone
	}
	season.UpdatedAt = time.Now().UTC()

	if err := s.seasonRepo.Update(ctx, season); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update season", err)
	}
	return season, nil
}
