package series

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// --- Mock SeriesRepository ---

type mockSeriesRepo struct {
	createFn         func(ctx context.Context, s *entity.Series) error
	getByIDFn        func(ctx context.Context, id uuid.UUID) (*entity.Series, error)
	getBySlugFn      func(ctx context.Context, slug valueobject.Slug) (*entity.Series, error)
	updateFn         func(ctx context.Context, s *entity.Series) error
	deleteFn         func(ctx context.Context, id uuid.UUID) error
	listFn           func(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, error)
	listPublishedFn  func(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, error)
	listByUploaderFn func(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Series, int64, error)
	existsBySlugFn   func(ctx context.Context, slug valueobject.Slug) (bool, error)
	setGenresFn      func(ctx context.Context, seriesID uuid.UUID, genreIDs []uuid.UUID) error
	getGenreIDsFn    func(ctx context.Context, seriesID uuid.UUID) ([]uuid.UUID, error)
}

func (m *mockSeriesRepo) Create(ctx context.Context, s *entity.Series) error {
	if m.createFn != nil {
		return m.createFn(ctx, s)
	}
	return nil
}

func (m *mockSeriesRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Series, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSeriesRepo) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Series, error) {
	if m.getBySlugFn != nil {
		return m.getBySlugFn(ctx, slug)
	}
	return nil, nil
}

func (m *mockSeriesRepo) Update(ctx context.Context, s *entity.Series) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, s)
	}
	return nil
}

func (m *mockSeriesRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockSeriesRepo) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, fs)
	}
	return nil, 0, nil
}

func (m *mockSeriesRepo) ListPublished(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, error) {
	if m.listPublishedFn != nil {
		return m.listPublishedFn(ctx, fs)
	}
	return nil, 0, nil
}

func (m *mockSeriesRepo) ListByUploader(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Series, int64, error) {
	if m.listByUploaderFn != nil {
		return m.listByUploaderFn(ctx, uploaderID, fs)
	}
	return nil, 0, nil
}

func (m *mockSeriesRepo) ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error) {
	if m.existsBySlugFn != nil {
		return m.existsBySlugFn(ctx, slug)
	}
	return false, nil
}

func (m *mockSeriesRepo) SetGenres(ctx context.Context, seriesID uuid.UUID, genreIDs []uuid.UUID) error {
	if m.setGenresFn != nil {
		return m.setGenresFn(ctx, seriesID, genreIDs)
	}
	return nil
}

func (m *mockSeriesRepo) GetGenreIDs(ctx context.Context, seriesID uuid.UUID) ([]uuid.UUID, error) {
	if m.getGenreIDsFn != nil {
		return m.getGenreIDsFn(ctx, seriesID)
	}
	return nil, nil
}

// --- Mock SeriesWatchProgressRepository ---

type mockSeriesWpRepo struct {
	upsertFn func(ctx context.Context, userID, seriesID, lastEpisodeID uuid.UUID) error
	getFn    func(ctx context.Context, userID, seriesID uuid.UUID) (*entity.SeriesWatchProgress, error)
	deleteFn func(ctx context.Context, userID, seriesID uuid.UUID) error
}

func (m *mockSeriesWpRepo) Upsert(ctx context.Context, userID, seriesID, lastEpisodeID uuid.UUID) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, userID, seriesID, lastEpisodeID)
	}
	return nil
}

func (m *mockSeriesWpRepo) Get(ctx context.Context, userID, seriesID uuid.UUID) (*entity.SeriesWatchProgress, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID, seriesID)
	}
	return nil, nil
}

func (m *mockSeriesWpRepo) Delete(ctx context.Context, userID, seriesID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, seriesID)
	}
	return nil
}

// --- Mock VideoRepository ---

type mockVideoRepo struct {
	getByIDFn          func(ctx context.Context, id uuid.UUID) (*entity.Video, error)
	setSeriesEpisodeFn func(ctx context.Context, videoID, seriesID uuid.UUID, season, episode int) error
	removeFromSeriesFn func(ctx context.Context, videoID uuid.UUID) error
	listBySeriesFn     func(ctx context.Context, seriesID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error)
}

func (m *mockVideoRepo) Create(context.Context, *entity.Video) error { return nil }
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
func (m *mockVideoRepo) ListByUploader(context.Context, uuid.UUID, *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}
func (m *mockVideoRepo) ExistsBySlug(context.Context, valueobject.Slug) (bool, error) {
	return false, nil
}
func (m *mockVideoRepo) SetGenres(context.Context, uuid.UUID, []uuid.UUID) error     { return nil }
func (m *mockVideoRepo) GetGenreIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) { return nil, nil }
func (m *mockVideoRepo) ListAllStorageKeys(context.Context) ([]repository.VideoStorageInfo, error) {
	return nil, nil
}

func (m *mockVideoRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockVideoRepo) SetSeriesEpisode(ctx context.Context, videoID, seriesID uuid.UUID, season, episode int) error {
	if m.setSeriesEpisodeFn != nil {
		return m.setSeriesEpisodeFn(ctx, videoID, seriesID, season, episode)
	}
	return nil
}

func (m *mockVideoRepo) RemoveFromSeries(ctx context.Context, videoID uuid.UUID) error {
	if m.removeFromSeriesFn != nil {
		return m.removeFromSeriesFn(ctx, videoID)
	}
	return nil
}

func (m *mockVideoRepo) ListBySeries(ctx context.Context, seriesID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	if m.listBySeriesFn != nil {
		return m.listBySeriesFn(ctx, seriesID, fs)
	}
	return nil, 0, nil
}
func (m *mockVideoRepo) ListFeatured(_ context.Context, _ *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}
func (m *mockVideoRepo) SetFeatured(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (m *mockVideoRepo) UnsetFeatured(_ context.Context, _ uuid.UUID) error         { return nil }

// --- Stub GenreRepository ---

type stubGenreRepo struct{}

func (stubGenreRepo) Create(context.Context, *entity.Genre) error               { return nil }
func (stubGenreRepo) GetByID(context.Context, uuid.UUID) (*entity.Genre, error) { return nil, nil }
func (stubGenreRepo) GetBySlug(context.Context, valueobject.Slug) (*entity.Genre, error) {
	return nil, nil
}
func (stubGenreRepo) Update(context.Context, *entity.Genre) error { return nil }
func (stubGenreRepo) Delete(context.Context, uuid.UUID) error     { return nil }
func (stubGenreRepo) List(context.Context, *filter.FilterSet) ([]entity.Genre, int64, error) {
	return nil, 0, nil
}
func (stubGenreRepo) ExistsBySlug(context.Context, valueobject.Slug) (bool, error) { return false, nil }

// --- Mock Dispatcher ---

type mockDispatcher struct {
	dispatched []event.Event
}

func (m *mockDispatcher) Dispatch(evt event.Event) error {
	m.dispatched = append(m.dispatched, evt)
	return nil
}

func (m *mockDispatcher) Subscribe(_ string, _ func(event.Event) error) {}
func (m *mockDispatcher) SubscribeAll(_ func(event.Event) error)        {}

// --- Fixtures ---

func newDraftSeries() *entity.Series {
	slug, _ := valueobject.NewSlug("test-series")
	s := entity.NewSeries("Test Series", slug.WithRandomID(), "desc", "creator", uuid.New())
	return s
}

func newPublishedSeries() *entity.Series {
	s := newDraftSeries()
	s.Publish()
	return s
}

func newPublishedVideo(seriesID uuid.UUID, season, episode int) *entity.Video {
	slug, _ := valueobject.NewSlug("test-episode")
	v := entity.NewVideo("Test Episode", slug.WithRandomID(), "desc", "creator", "PG", "videos/x/original.mp4", "video/mp4", 1024, uuid.New())
	v.MarkTranscoded()
	v.SeriesID = &seriesID
	v.SeasonNumber = &season
	v.EpisodeNumber = &episode
	return v
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestService(sr *mockSeriesRepo, swpr *mockSeriesWpRepo, vr *mockVideoRepo, d *mockDispatcher) *Service {
	if sr == nil {
		sr = &mockSeriesRepo{}
	}
	if swpr == nil {
		swpr = &mockSeriesWpRepo{}
	}
	if vr == nil {
		vr = &mockVideoRepo{}
	}
	if d == nil {
		d = &mockDispatcher{}
	}
	return NewService(sr, swpr, vr, stubGenreRepo{}, d, newTestLogger())
}
