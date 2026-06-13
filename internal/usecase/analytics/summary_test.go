package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestGetSummary(t *testing.T) {
	uploaderID := uuid.New()

	t.Run("publisher scope is passed through to all queries", func(t *testing.T) {
		var totalsScopes []repository.PlaybackScope
		var topVideosUploader *uuid.UUID
		repo := &mockAnalyticsRepo{
			getTotalsFn: func(_ context.Context, scope repository.PlaybackScope, _, _ time.Time) (*repository.PlaybackTotals, error) {
				totalsScopes = append(totalsScopes, scope)
				return &repository.PlaybackTotals{Views: 10}, nil
			},
			getTopVideosFn: func(_ context.Context, uploader *uuid.UUID, _, _ time.Time, limit int) ([]repository.VideoRollup, error) {
				topVideosUploader = uploader
				if limit != topVideosLimit {
					t.Errorf("limit = %d, want %d", limit, topVideosLimit)
				}
				return []repository.VideoRollup{{Title: "Top"}}, nil
			},
		}
		svc := newTestService(repo, nil, nil)

		out, appErr := svc.GetSummary(context.Background(), &SummaryInput{UploaderID: &uploaderID, RangeKey: "30d"})
		if appErr != nil {
			t.Fatalf("unexpected error: %v", appErr)
		}

		if len(totalsScopes) != 2 {
			t.Fatalf("GetTotals called %d times, want 2 (current + previous)", len(totalsScopes))
		}
		for _, scope := range totalsScopes {
			if scope.UploaderID == nil || *scope.UploaderID != uploaderID {
				t.Errorf("scope uploader = %v, want %s", scope.UploaderID, uploaderID)
			}
		}
		if topVideosUploader == nil || *topVideosUploader != uploaderID {
			t.Errorf("top videos uploader = %v, want %s", topVideosUploader, uploaderID)
		}
		if len(out.TopVideos) != 1 || out.TopVideos[0].Title != "Top" {
			t.Errorf("top videos = %+v", out.TopVideos)
		}
	})

	t.Run("nil uploader produces platform-wide scope", func(t *testing.T) {
		repo := &mockAnalyticsRepo{
			getTotalsFn: func(_ context.Context, scope repository.PlaybackScope, _, _ time.Time) (*repository.PlaybackTotals, error) {
				if scope.UploaderID != nil || scope.VideoID != nil {
					t.Errorf("expected unscoped query, got %+v", scope)
				}
				return &repository.PlaybackTotals{}, nil
			},
			getTopVideosFn: func(_ context.Context, uploader *uuid.UUID, _, _ time.Time, _ int) ([]repository.VideoRollup, error) {
				if uploader != nil {
					t.Errorf("expected nil uploader, got %s", *uploader)
				}
				return nil, nil
			},
		}
		svc := newTestService(repo, nil, nil)

		if _, appErr := svc.GetSummary(context.Background(), &SummaryInput{RangeKey: "7d"}); appErr != nil {
			t.Fatalf("unexpected error: %v", appErr)
		}
	})

	t.Run("timeseries is gap-filled across the window", func(t *testing.T) {
		repo := &mockAnalyticsRepo{
			getDailySeriesFn: func(context.Context, repository.PlaybackScope, time.Time, time.Time) ([]repository.DailyStat, error) {
				return []repository.DailyStat{
					{Day: fixedNow.AddDate(0, 0, -3), Views: 5, WatchedSeconds: 300},
				}, nil
			},
		}
		svc := newTestService(repo, nil, nil)

		out, appErr := svc.GetSummary(context.Background(), &SummaryInput{UploaderID: &uploaderID, RangeKey: "7d"})
		if appErr != nil {
			t.Fatalf("unexpected error: %v", appErr)
		}
		if len(out.Timeseries) != 8 {
			t.Fatalf("timeseries length = %d, want 8 (7 days + today)", len(out.Timeseries))
		}
		var nonZero int
		for _, d := range out.Timeseries {
			if d.Views > 0 {
				nonZero++
				if d.Views != 5 || d.WatchedSeconds != 300 {
					t.Errorf("data day = %+v", d)
				}
			}
		}
		if nonZero != 1 {
			t.Errorf("non-zero days = %d, want 1", nonZero)
		}
	})

	t.Run("invalid range", func(t *testing.T) {
		svc := newTestService(nil, nil, nil)
		_, appErr := svc.GetSummary(context.Background(), &SummaryInput{RangeKey: "1y"})
		if appErr == nil || appErr.Code != apperror.CodeInvalidAnalyticsRange {
			t.Fatalf("expected %s, got %v", apperror.CodeInvalidAnalyticsRange, appErr)
		}
	})

	t.Run("repo error surfaces as internal", func(t *testing.T) {
		repo := &mockAnalyticsRepo{
			getTotalsFn: func(context.Context, repository.PlaybackScope, time.Time, time.Time) (*repository.PlaybackTotals, error) {
				return nil, errors.New("db down")
			},
		}
		svc := newTestService(repo, nil, nil)
		_, appErr := svc.GetSummary(context.Background(), &SummaryInput{RangeKey: "30d"})
		if appErr == nil || appErr.Code != apperror.CodeInternalError {
			t.Fatalf("expected %s, got %v", apperror.CodeInternalError, appErr)
		}
	})
}
