package video

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// --- Mock VideoRepository ---

type mockVideoRepo struct {
	createFn        func(ctx context.Context, video *entity.Video) error
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*entity.Video, error)
	getBySlugFn     func(ctx context.Context, slug valueobject.Slug) (*entity.Video, error)
	updateFn        func(ctx context.Context, video *entity.Video) error
	deleteFn        func(ctx context.Context, id uuid.UUID) error
	listFn          func(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error)
	listPublishedFn func(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error)
	existsBySlugFn  func(ctx context.Context, slug valueobject.Slug) (bool, error)
	setGenresFn     func(ctx context.Context, videoID uuid.UUID, genreIDs []uuid.UUID) error
	getGenreIDsFn   func(ctx context.Context, videoID uuid.UUID) ([]uuid.UUID, error)
}

func (m *mockVideoRepo) Create(ctx context.Context, v *entity.Video) error {
	if m.createFn != nil {
		return m.createFn(ctx, v)
	}
	return nil
}

func (m *mockVideoRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockVideoRepo) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Video, error) {
	if m.getBySlugFn != nil {
		return m.getBySlugFn(ctx, slug)
	}
	return nil, nil
}

func (m *mockVideoRepo) Update(ctx context.Context, v *entity.Video) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, v)
	}
	return nil
}

func (m *mockVideoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockVideoRepo) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, fs)
	}
	return nil, 0, nil
}

func (m *mockVideoRepo) ListPublished(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	if m.listPublishedFn != nil {
		return m.listPublishedFn(ctx, fs)
	}
	return nil, 0, nil
}

func (m *mockVideoRepo) ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error) {
	if m.existsBySlugFn != nil {
		return m.existsBySlugFn(ctx, slug)
	}
	return false, nil
}

func (m *mockVideoRepo) SetGenres(ctx context.Context, videoID uuid.UUID, genreIDs []uuid.UUID) error {
	if m.setGenresFn != nil {
		return m.setGenresFn(ctx, videoID, genreIDs)
	}
	return nil
}

func (m *mockVideoRepo) GetGenreIDs(ctx context.Context, videoID uuid.UUID) ([]uuid.UUID, error) {
	if m.getGenreIDsFn != nil {
		return m.getGenreIDsFn(ctx, videoID)
	}
	return nil, nil
}

// --- Stub repos for fields the callback tests don't use ---

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

type stubSubRepo struct{}

func (stubSubRepo) Create(context.Context, *entity.Subscription) error { return nil }
func (stubSubRepo) GetByID(context.Context, uuid.UUID) (*entity.Subscription, error) {
	return nil, nil
}
func (stubSubRepo) GetActiveByUserID(context.Context, uuid.UUID) (*entity.Subscription, error) {
	return nil, nil
}
func (stubSubRepo) Update(context.Context, *entity.Subscription) error { return nil }
func (stubSubRepo) Cancel(context.Context, uuid.UUID) error            { return nil }
func (stubSubRepo) ListByUserID(context.Context, uuid.UUID) ([]entity.Subscription, error) {
	return nil, nil
}

type stubPlanRepo struct{}

func (stubPlanRepo) Create(context.Context, *entity.Plan) error               { return nil }
func (stubPlanRepo) GetByID(context.Context, uuid.UUID) (*entity.Plan, error) { return nil, nil }
func (stubPlanRepo) GetBySlug(context.Context, valueobject.Slug) (*entity.Plan, error) {
	return nil, nil
}
func (stubPlanRepo) Update(context.Context, *entity.Plan) error { return nil }
func (stubPlanRepo) Delete(context.Context, uuid.UUID) error    { return nil }
func (stubPlanRepo) List(context.Context, *filter.FilterSet) ([]entity.Plan, int64, error) {
	return nil, 0, nil
}
func (stubPlanRepo) ListActive(context.Context) ([]entity.Plan, error)            { return nil, nil }
func (stubPlanRepo) ExistsBySlug(context.Context, valueobject.Slug) (bool, error) { return false, nil }

type stubStorage struct{}

func (stubStorage) Upload(context.Context, string, io.Reader, string) (string, error) { return "", nil }
func (stubStorage) Delete(context.Context, string) error                              { return nil }
func (stubStorage) Move(context.Context, string, string) error                        { return nil }
func (stubStorage) PublicURL(string) string                                           { return "" }
func (stubStorage) PresignedGetURL(context.Context, string, time.Duration) (string, error) {
	return "", nil
}

// --- Mock Dispatcher ---

type mockDispatcher struct {
	dispatched []event.Event
}

func (m *mockDispatcher) Dispatch(evt event.Event) error {
	m.dispatched = append(m.dispatched, evt)
	return nil
}

func (m *mockDispatcher) Subscribe(_ string, _ func(event.Event) error) {}

// --- Mock CDNClient ---

type mockCDNClient struct {
	calls       []string
	triggerErr  error
	triggeredCh chan string
}

func (m *mockCDNClient) TriggerTranscode(videoID string) error {
	m.calls = append(m.calls, videoID)
	if m.triggeredCh != nil {
		m.triggeredCh <- videoID
	}
	return m.triggerErr
}

// --- Test Helpers ---

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestService(repo *mockVideoRepo, dispatcher *mockDispatcher, opts ...Option) *Service {
	if repo == nil {
		repo = &mockVideoRepo{}
	}
	if dispatcher == nil {
		dispatcher = &mockDispatcher{}
	}
	return NewService(repo, stubGenreRepo{}, stubSubRepo{}, stubPlanRepo{}, stubStorage{}, dispatcher, newTestLogger(), opts...)
}
