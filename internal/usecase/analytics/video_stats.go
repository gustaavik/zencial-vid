package analytics

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// VideoStatsInput identifies the video, caller, and reporting window for
// per-video analytics.
type VideoStatsInput struct {
	VideoID     uuid.UUID
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
	RangeKey    string
}

// VideoStatsOutput is the full per-video analytics report.
type VideoStatsOutput struct {
	VideoID    uuid.UUID
	Range      valueobject.AnalyticsRange
	Totals     repository.PlaybackTotals
	Deltas     *Deltas
	Timeseries []repository.DailyStat
	Retention  []float64
	Sources    []repository.BreakdownItem
	Countries  []repository.BreakdownItem
	Platforms  []repository.BreakdownItem
}

// GetVideoStats returns viewing statistics for a single video over a range.
// Publishers may only query stats for videos they uploaded; admins may query
// any video.
func (s *Service) GetVideoStats(ctx context.Context, in *VideoStatsInput) (*VideoStatsOutput, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, in.VideoID)
	if err != nil {
		s.log.Error("analytics: getting video", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if !entity.HasRole(in.CallerRoles, entity.RoleAdmin) && video.UploadedBy != in.CallerID {
		return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
	}

	rng, err := valueobject.NewAnalyticsRange(in.RangeKey, s.clock.Now())
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeInvalidAnalyticsRange, "range must be one of 7d, 30d, 90d, all", err)
	}

	videoID := in.VideoID
	scope := repository.PlaybackScope{VideoID: &videoID}

	totals, err := s.analyticsRepo.GetTotals(ctx, scope, rng.From, rng.To)
	if err != nil {
		s.log.Error("analytics: getting video totals", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video stats", err)
	}

	var deltas *Deltas
	if rng.HasPrev {
		prev, err := s.analyticsRepo.GetTotals(ctx, scope, rng.PrevFrom, rng.PrevTo)
		if err != nil {
			s.log.Error("analytics: getting previous video totals", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video stats", err)
		}
		deltas = computeDeltas(totals, prev)
	}

	series, err := s.analyticsRepo.GetDailySeries(ctx, scope, rng.From, rng.To)
	if err != nil {
		s.log.Error("analytics: getting video daily series", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video stats", err)
	}
	if rng.HasPrev {
		series = fillDailyGaps(series, rng.From, rng.To)
	}

	retention, err := s.analyticsRepo.GetRetention(ctx, in.VideoID, rng.From, rng.To)
	if err != nil {
		s.log.Error("analytics: getting retention curve", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video stats", err)
	}

	out := &VideoStatsOutput{
		VideoID:    in.VideoID,
		Range:      rng,
		Totals:     *totals,
		Deltas:     deltas,
		Timeseries: series,
		Retention:  retention,
	}

	breakdowns := []struct {
		dim  repository.BreakdownDimension
		dest *[]repository.BreakdownItem
	}{
		{repository.BreakdownSource, &out.Sources},
		{repository.BreakdownCountry, &out.Countries},
		{repository.BreakdownPlatform, &out.Platforms},
	}
	for _, b := range breakdowns {
		items, err := s.analyticsRepo.GetBreakdown(ctx, in.VideoID, b.dim, rng.From, rng.To)
		if err != nil {
			s.log.Error("analytics: getting breakdown", "dimension", b.dim, "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video stats", err)
		}
		*b.dest = items
	}

	return out, nil
}
