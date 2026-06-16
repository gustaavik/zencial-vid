package analytics

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

var fixedNow = time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)

// --- Mock AnalyticsRepository ---

type mockAnalyticsRepo struct {
	getTotalsFn      func(ctx context.Context, scope repository.PlaybackScope, from, to time.Time) (*repository.PlaybackTotals, error)
	getDailySeriesFn func(ctx context.Context, scope repository.PlaybackScope, from, to time.Time) ([]repository.DailyStat, error)
	getTopVideosFn   func(ctx context.Context, uploaderID *uuid.UUID, from, to time.Time, limit int) ([]repository.VideoRollup, error)
	getRetentionFn   func(ctx context.Context, videoID uuid.UUID, from, to time.Time) ([]float64, error)
	getBreakdownFn   func(ctx context.Context, videoID uuid.UUID, dim repository.BreakdownDimension, from, to time.Time) ([]repository.BreakdownItem, error)
}

func (m *mockAnalyticsRepo) GetTotals(ctx context.Context, scope repository.PlaybackScope, from, to time.Time) (*repository.PlaybackTotals, error) {
	if m.getTotalsFn != nil {
		return m.getTotalsFn(ctx, scope, from, to)
	}
	return &repository.PlaybackTotals{}, nil
}

func (m *mockAnalyticsRepo) GetDailySeries(ctx context.Context, scope repository.PlaybackScope, from, to time.Time) ([]repository.DailyStat, error) {
	if m.getDailySeriesFn != nil {
		return m.getDailySeriesFn(ctx, scope, from, to)
	}
	return nil, nil
}

func (m *mockAnalyticsRepo) GetTopVideos(ctx context.Context, uploaderID *uuid.UUID, from, to time.Time, limit int) ([]repository.VideoRollup, error) {
	if m.getTopVideosFn != nil {
		return m.getTopVideosFn(ctx, uploaderID, from, to, limit)
	}
	return nil, nil
}

func (m *mockAnalyticsRepo) GetRetention(ctx context.Context, videoID uuid.UUID, from, to time.Time) ([]float64, error) {
	if m.getRetentionFn != nil {
		return m.getRetentionFn(ctx, videoID, from, to)
	}
	return make([]float64, entity.RetentionBuckets), nil
}

func (m *mockAnalyticsRepo) GetBreakdown(ctx context.Context, videoID uuid.UUID, dim repository.BreakdownDimension, from, to time.Time) ([]repository.BreakdownItem, error) {
	if m.getBreakdownFn != nil {
		return m.getBreakdownFn(ctx, videoID, dim, from, to)
	}
	return nil, nil
}

// --- Mock PlaybackSessionRepository ---

type mockPlaybackRepo struct {
	upsertHeartbeatFn func(ctx context.Context, hb *repository.PlaybackHeartbeat) error
}

func (m *mockPlaybackRepo) UpsertHeartbeat(ctx context.Context, hb *repository.PlaybackHeartbeat) error {
	if m.upsertHeartbeatFn != nil {
		return m.upsertHeartbeatFn(ctx, hb)
	}
	return nil
}

// --- Mock VideoRepository (only GetByID is exercised by analytics) ---

type mockVideoRepo struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*entity.Video, error)
}

func (m *mockVideoRepo) Create(context.Context, *entity.Video) error { return nil }
func (m *mockVideoRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockVideoRepo) GetBySlug(context.Context, valueobject.Slug) (*entity.Video, error) {
	return nil, nil
}
func (m *mockVideoRepo) Update(context.Context, *entity.Video) error { return nil }
func (m *mockVideoRepo) Delete(context.Context, uuid.UUID) error     { return nil }
func (m *mockVideoRepo) List(context.Context, *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}

func (m *mockVideoRepo) ListPublished(context.Context, *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}

func (m *mockVideoRepo) ListAdmin(context.Context, *filter.FilterSet, *uuid.UUID) ([]entity.Video, int64, error) {
	return nil, 0, nil
}

func (m *mockVideoRepo) Stats(context.Context) (*repository.VideoStats, error) {
	return &repository.VideoStats{ByStatus: map[string]int64{}, BySubmission: map[string]int64{}}, nil
}

func (m *mockVideoRepo) ListByUploader(context.Context, uuid.UUID, *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}

func (m *mockVideoRepo) ExistsBySlug(context.Context, valueobject.Slug) (bool, error) {
	return false, nil
}

func (m *mockVideoRepo) SetGenres(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (m *mockVideoRepo) GetGenreIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (m *mockVideoRepo) ListAllStorageKeys(context.Context) ([]repository.VideoStorageInfo, error) {
	return nil, nil
}

func (m *mockVideoRepo) SetSeriesEpisode(context.Context, uuid.UUID, uuid.UUID, int, int) error {
	return nil
}
func (m *mockVideoRepo) RemoveFromSeries(context.Context, uuid.UUID) error { return nil }
func (m *mockVideoRepo) ListBySeries(context.Context, uuid.UUID, *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}

func (m *mockVideoRepo) ListFeatured(context.Context, *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}
func (m *mockVideoRepo) SetFeatured(context.Context, uuid.UUID, string) error { return nil }
func (m *mockVideoRepo) UnsetFeatured(context.Context, uuid.UUID) error       { return nil }

// --- Test helpers ---

func newTestService(analyticsRepo *mockAnalyticsRepo, playbackRepo *mockPlaybackRepo, videoRepo *mockVideoRepo) *Service {
	if analyticsRepo == nil {
		analyticsRepo = &mockAnalyticsRepo{}
	}
	if playbackRepo == nil {
		playbackRepo = &mockPlaybackRepo{}
	}
	if videoRepo == nil {
		videoRepo = &mockVideoRepo{}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(analyticsRepo, playbackRepo, videoRepo, log, clock.MockClock{FixedTime: fixedNow})
}

func testVideo(id, uploadedBy uuid.UUID, durationSeconds int64) *entity.Video {
	return &entity.Video{
		ID:         id,
		Title:      "Test Video",
		Duration:   valueobject.NewDuration(durationSeconds),
		UploadedBy: uploadedBy,
	}
}
