package analytics

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// topVideosLimit caps the dashboard "top videos" table size.
const topVideosLimit = 10

// SummaryInput selects the scope and reporting window for a dashboard summary.
// A nil UploaderID produces a platform-wide summary (admin only — enforced by
// routing).
type SummaryInput struct {
	UploaderID *uuid.UUID
	RangeKey   string
}

// SummaryOutput is the dashboard summary report.
type SummaryOutput struct {
	Range      valueobject.AnalyticsRange
	Totals     repository.PlaybackTotals
	Deltas     *Deltas
	Timeseries []repository.DailyStat
	TopVideos  []repository.VideoRollup
}

// GetSummary returns aggregate viewing statistics over a range, scoped to one
// uploader's videos or (with a nil uploader) the whole platform.
func (s *Service) GetSummary(ctx context.Context, in *SummaryInput) (*SummaryOutput, *apperror.AppError) {
	rng, err := valueobject.NewAnalyticsRange(in.RangeKey, s.clock.Now())
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeInvalidAnalyticsRange, "range must be one of 7d, 30d, 90d, all", err)
	}

	scope := repository.PlaybackScope{UploaderID: in.UploaderID}

	totals, err := s.analyticsRepo.GetTotals(ctx, scope, rng.From, rng.To)
	if err != nil {
		s.log.Error("analytics: getting summary totals", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get analytics summary", err)
	}

	var deltas *Deltas
	if rng.HasPrev {
		prev, err := s.analyticsRepo.GetTotals(ctx, scope, rng.PrevFrom, rng.PrevTo)
		if err != nil {
			s.log.Error("analytics: getting previous summary totals", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to get analytics summary", err)
		}
		deltas = computeDeltas(totals, prev)
	}

	series, err := s.analyticsRepo.GetDailySeries(ctx, scope, rng.From, rng.To)
	if err != nil {
		s.log.Error("analytics: getting summary daily series", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get analytics summary", err)
	}
	if rng.HasPrev {
		series = fillDailyGaps(series, rng.From, rng.To)
	}

	topVideos, err := s.analyticsRepo.GetTopVideos(ctx, in.UploaderID, rng.From, rng.To, topVideosLimit)
	if err != nil {
		s.log.Error("analytics: getting top videos", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get analytics summary", err)
	}

	return &SummaryOutput{
		Range:      rng,
		Totals:     *totals,
		Deltas:     deltas,
		Timeseries: series,
		TopVideos:  topVideos,
	}, nil
}
