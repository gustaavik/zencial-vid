package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestGetVideoStats(t *testing.T) {
	videoID := uuid.New()
	ownerID := uuid.New()
	strangerID := uuid.New()

	curTotals := &repository.PlaybackTotals{Views: 200, WatchedSeconds: 6000, UniqueViewers: 150, AvgPercentWatched: 60, FinishRate: 40}
	prevTotals := &repository.PlaybackTotals{Views: 100, WatchedSeconds: 3000, UniqueViewers: 100, AvgPercentWatched: 50, FinishRate: 50}

	totalsByWindow := func(cur, prev *repository.PlaybackTotals) func(ctx context.Context, scope repository.PlaybackScope, from, to time.Time) (*repository.PlaybackTotals, error) {
		return func(_ context.Context, _ repository.PlaybackScope, _, to time.Time) (*repository.PlaybackTotals, error) {
			if to.Equal(fixedNow) {
				return cur, nil
			}
			return prev, nil
		}
	}

	tests := []struct {
		name     string
		input    *VideoStatsInput
		repo     *mockAnalyticsRepo
		wantCode string
		check    func(t *testing.T, out *VideoStatsOutput)
	}{
		{
			name:  "owner gets stats with deltas",
			input: &VideoStatsInput{VideoID: videoID, CallerID: ownerID, RangeKey: "30d"},
			repo:  &mockAnalyticsRepo{getTotalsFn: totalsByWindow(curTotals, prevTotals)},
			check: func(t *testing.T, out *VideoStatsOutput) {
				if out.Totals != *curTotals {
					t.Errorf("totals = %+v, want %+v", out.Totals, *curTotals)
				}
				if out.Deltas == nil {
					t.Fatal("expected deltas")
				}
				if out.Deltas.ViewsPct != 100 {
					t.Errorf("views delta = %v, want 100", out.Deltas.ViewsPct)
				}
				if out.Deltas.FinishRatePts != -10 {
					t.Errorf("finish rate delta = %v, want -10 points", out.Deltas.FinishRatePts)
				}
				if len(out.Retention) != entity.RetentionBuckets {
					t.Errorf("retention length = %d, want %d", len(out.Retention), entity.RetentionBuckets)
				}
				if len(out.Timeseries) != 31 {
					t.Errorf("timeseries length = %d, want 31 gap-filled days", len(out.Timeseries))
				}
			},
		},
		{
			name:  "zero previous views omits deltas",
			input: &VideoStatsInput{VideoID: videoID, CallerID: ownerID, RangeKey: "7d"},
			repo:  &mockAnalyticsRepo{getTotalsFn: totalsByWindow(curTotals, &repository.PlaybackTotals{})},
			check: func(t *testing.T, out *VideoStatsOutput) {
				if out.Deltas != nil {
					t.Errorf("expected nil deltas, got %+v", out.Deltas)
				}
			},
		},
		{
			name:  "all range skips previous window and gap fill",
			input: &VideoStatsInput{VideoID: videoID, CallerID: ownerID, RangeKey: "all"},
			repo: &mockAnalyticsRepo{
				getTotalsFn: func(_ context.Context, _ repository.PlaybackScope, _, to time.Time) (*repository.PlaybackTotals, error) {
					if !to.Equal(fixedNow) {
						t.Error("previous window should not be queried for range=all")
					}
					return curTotals, nil
				},
				getDailySeriesFn: func(context.Context, repository.PlaybackScope, time.Time, time.Time) ([]repository.DailyStat, error) {
					return []repository.DailyStat{{Day: fixedNow, Views: 1}}, nil
				},
			},
			check: func(t *testing.T, out *VideoStatsOutput) {
				if out.Deltas != nil {
					t.Errorf("expected nil deltas for range=all, got %+v", out.Deltas)
				}
				if len(out.Timeseries) != 1 {
					t.Errorf("timeseries length = %d, want 1 (no gap fill for all)", len(out.Timeseries))
				}
			},
		},
		{
			name:  "breakdowns are mapped per dimension",
			input: &VideoStatsInput{VideoID: videoID, CallerID: ownerID, RangeKey: "30d"},
			repo: &mockAnalyticsRepo{
				getBreakdownFn: func(_ context.Context, _ uuid.UUID, dim repository.BreakdownDimension, _, _ time.Time) ([]repository.BreakdownItem, error) {
					return []repository.BreakdownItem{{Key: string(dim), Views: 1}}, nil
				},
			},
			check: func(t *testing.T, out *VideoStatsOutput) {
				if len(out.Sources) != 1 || out.Sources[0].Key != string(repository.BreakdownSource) {
					t.Errorf("sources = %+v", out.Sources)
				}
				if len(out.Countries) != 1 || out.Countries[0].Key != string(repository.BreakdownCountry) {
					t.Errorf("countries = %+v", out.Countries)
				}
				if len(out.Platforms) != 1 || out.Platforms[0].Key != string(repository.BreakdownPlatform) {
					t.Errorf("platforms = %+v", out.Platforms)
				}
			},
		},
		{
			name: "admin bypasses ownership",
			input: &VideoStatsInput{
				VideoID:     videoID,
				CallerID:    strangerID,
				CallerRoles: []entity.UserRole{entity.RoleAdmin},
				RangeKey:    "30d",
			},
			repo: &mockAnalyticsRepo{},
		},
		{
			name:     "non-owner is forbidden",
			input:    &VideoStatsInput{VideoID: videoID, CallerID: strangerID, RangeKey: "30d"},
			repo:     &mockAnalyticsRepo{},
			wantCode: apperror.CodeVideoOwnershipRequired,
		},
		{
			name:     "invalid range",
			input:    &VideoStatsInput{VideoID: videoID, CallerID: ownerID, RangeKey: "14d"},
			repo:     &mockAnalyticsRepo{},
			wantCode: apperror.CodeInvalidAnalyticsRange,
		},
		{
			name:  "repo error surfaces as internal",
			input: &VideoStatsInput{VideoID: videoID, CallerID: ownerID, RangeKey: "30d"},
			repo: &mockAnalyticsRepo{
				getTotalsFn: func(context.Context, repository.PlaybackScope, time.Time, time.Time) (*repository.PlaybackTotals, error) {
					return nil, errors.New("db down")
				},
			},
			wantCode: apperror.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			videoRepo := &mockVideoRepo{
				getByIDFn: func(context.Context, uuid.UUID) (*entity.Video, error) {
					return testVideo(videoID, ownerID, 600), nil
				},
			}
			svc := newTestService(tt.repo, nil, videoRepo)

			out, appErr := svc.GetVideoStats(context.Background(), tt.input)

			if tt.wantCode != "" {
				if appErr == nil || appErr.Code != tt.wantCode {
					t.Fatalf("error = %v, want code %s", appErr, tt.wantCode)
				}
				return
			}
			if appErr != nil {
				t.Fatalf("unexpected error: %v", appErr)
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}

func TestGetVideoStatsVideoNotFound(t *testing.T) {
	videoRepo := &mockVideoRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*entity.Video, error) { return nil, nil },
	}
	svc := newTestService(nil, nil, videoRepo)

	_, appErr := svc.GetVideoStats(context.Background(), &VideoStatsInput{VideoID: uuid.New(), CallerID: uuid.New(), RangeKey: "30d"})
	if appErr == nil || appErr.Code != apperror.CodeVideoNotFound {
		t.Fatalf("expected %s, got %v", apperror.CodeVideoNotFound, appErr)
	}
}
