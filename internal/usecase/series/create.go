package series

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

// CreateInput holds the data needed to create a series.
type CreateInput struct {
	Title               string
	Description         string
	Creator             string
	SeriesType          string
	Logline             string
	PrimaryLanguage     string
	OriginCountry       string
	ContentRating       string
	CoverImageKey       string
	PosterKey           string
	BannerKey           string
	TitleLogoKey        string
	UploadedBy          uuid.UUID
	GenreIDs            []uuid.UUID
	MinimumPlanLevel    *int
	AutoplayNext        *bool
	BingeMode           *bool
	HideEpisodeCount    *bool
	DefaultVisibility   string
	DefaultMonetization []string
}

// CreateOutput holds the result of a Create operation.
type CreateOutput struct {
	Series *entity.Series
}

// Create creates a new series in draft status.
func (s *Service) Create(ctx context.Context, input *CreateInput) (*CreateOutput, *apperror.AppError) {
	slug, err := valueobject.NewSlug(input.Title)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid title for slug", err)
	}
	slug = slug.WithRandomID()

	exists, err := s.seriesRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		s.log.Error("checking series slug existence", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check slug", err)
	}
	if exists {
		return nil, apperror.Conflict(apperror.CodeSeriesSlugConflict, "slug already exists", domain.ErrSeriesSlugExists)
	}

	series := entity.NewSeries(input.Title, slug, input.Description, input.Creator, input.UploadedBy)
	series.CoverImageKey = input.CoverImageKey
	series.PosterKey = input.PosterKey
	series.BannerKey = input.BannerKey
	series.TitleLogoKey = input.TitleLogoKey
	series.ContentRating = input.ContentRating
	series.OriginCountry = input.OriginCountry
	series.MinimumPlanLevel = input.MinimumPlanLevel

	if input.SeriesType != "" {
		series.SeriesType = entity.SeriesType(input.SeriesType)
	}
	if input.Logline != "" {
		series.Logline = input.Logline
	}
	if input.PrimaryLanguage != "" {
		series.PrimaryLanguage = input.PrimaryLanguage
	}
	if input.DefaultVisibility != "" {
		series.DefaultVisibility = entity.VideoVisibility(input.DefaultVisibility)
	}
	if input.DefaultMonetization != nil {
		series.DefaultMonetization = input.DefaultMonetization
	}
	if input.AutoplayNext != nil {
		series.AutoplayNext = *input.AutoplayNext
	}
	if input.BingeMode != nil {
		series.BingeMode = *input.BingeMode
	}
	if input.HideEpisodeCount != nil {
		series.HideEpisodeCount = *input.HideEpisodeCount
	}

	if err := s.seriesRepo.Create(ctx, series); err != nil {
		s.log.Error("creating series", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create series", err)
	}

	if len(input.GenreIDs) > 0 {
		if err := s.seriesRepo.SetGenres(ctx, series.ID, input.GenreIDs); err != nil {
			s.log.Error("setting series genres", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to set series genres", err)
		}
		series.GenreIDs = input.GenreIDs
	}

	if err := s.dispatcher.Dispatch(event.SeriesCreated{
		SeriesID:  series.ID,
		Title:     series.Title,
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching series created event", "error", err)
	}

	return &CreateOutput{Series: series}, nil
}
