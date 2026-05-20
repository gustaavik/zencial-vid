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
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// --- Mock VideoRepository ---

type mockVideoRepo struct {
	createFn             func(ctx context.Context, video *entity.Video) error
	getByIDFn            func(ctx context.Context, id uuid.UUID) (*entity.Video, error)
	getBySlugFn          func(ctx context.Context, slug valueobject.Slug) (*entity.Video, error)
	updateFn             func(ctx context.Context, video *entity.Video) error
	deleteFn             func(ctx context.Context, id uuid.UUID) error
	listFn               func(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error)
	listPublishedFn      func(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error)
	existsBySlugFn       func(ctx context.Context, slug valueobject.Slug) (bool, error)
	setGenresFn          func(ctx context.Context, videoID uuid.UUID, genreIDs []uuid.UUID) error
	getGenreIDsFn        func(ctx context.Context, videoID uuid.UUID) ([]uuid.UUID, error)
	listAllStorageKeysFn func(ctx context.Context) ([]repository.VideoStorageInfo, error)
	setSeriesEpisodeFn   func(ctx context.Context, videoID, seriesID uuid.UUID, season, episode int) error
	removeFromSeriesFn   func(ctx context.Context, videoID uuid.UUID) error
	listBySeriesFn       func(ctx context.Context, seriesID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error)
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

func (m *mockVideoRepo) ListAllStorageKeys(ctx context.Context) ([]repository.VideoStorageInfo, error) {
	if m.listAllStorageKeysFn != nil {
		return m.listAllStorageKeysFn(ctx)
	}
	return nil, nil
}

func (m *mockVideoRepo) ListByUploader(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
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
func (stubSubRepo) GetByStripeSubscriptionID(context.Context, string) (*entity.Subscription, error) {
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

type stubStorage struct {
	uploadFn        func(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	deleteFn        func(ctx context.Context, key string) error
	presignPutFn    func(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)
	statFn          func(ctx context.Context, key string) (*storage.ObjectInfo, error)
	listObjectsFn   func(ctx context.Context, prefix string) ([]string, error)
	deletedKeys     []string
	presignPutCalls []presignPutCall
}

type presignPutCall struct {
	Key         string
	ContentType string
	Expiry      time.Duration
}

func (s *stubStorage) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	if s.uploadFn != nil {
		return s.uploadFn(ctx, key, body, contentType)
	}
	return "", nil
}

func (s *stubStorage) Delete(ctx context.Context, key string) error {
	s.deletedKeys = append(s.deletedKeys, key)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, key)
	}
	return nil
}

func (s *stubStorage) Move(context.Context, string, string) error { return nil }

func (s *stubStorage) PublicURL(string) string { return "" }

func (s *stubStorage) PresignedGetURL(context.Context, string, time.Duration) (string, error) {
	return "", nil
}

func (s *stubStorage) PresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	s.presignPutCalls = append(s.presignPutCalls, presignPutCall{Key: key, ContentType: contentType, Expiry: expiry})
	if s.presignPutFn != nil {
		return s.presignPutFn(ctx, key, contentType, expiry)
	}
	return "https://stub-storage.example/" + key, nil
}

func (s *stubStorage) Stat(ctx context.Context, key string) (*storage.ObjectInfo, error) {
	if s.statFn != nil {
		return s.statFn(ctx, key)
	}
	return nil, nil
}

func (s *stubStorage) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	if s.listObjectsFn != nil {
		return s.listObjectsFn(ctx, prefix)
	}
	return nil, nil
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

func (m *mockDispatcher) SubscribeAll(_ func(event.Event) error) {}

// --- Mock CDNClient ---

type mockCDNClient struct {
	calls       []string
	triggerErr  error
	triggeredCh chan string

	// Upload-related fields, populated when the corresponding method is called.
	signedVideoURL           string
	signedVideoErr           error
	signVideoCalls           []signVideoCall
	uploadedThumbnails       []uploadedThumbnail
	uploadThumbnailErr       error
	uploadThumbnailObjectKey string
}

type signVideoCall struct {
	videoID  string
	filename string
	expiry   time.Duration
}

type uploadedThumbnail struct {
	videoID     string
	ext         string
	contentType string
	body        []byte
}

func (m *mockCDNClient) TriggerTranscode(videoID string) error {
	m.calls = append(m.calls, videoID)
	if m.triggeredCh != nil {
		m.triggeredCh <- videoID
	}
	return m.triggerErr
}

func (m *mockCDNClient) SignVideoUploadURL(videoID, filename string, expiry time.Duration) (string, time.Time, error) {
	m.signVideoCalls = append(m.signVideoCalls, signVideoCall{videoID: videoID, filename: filename, expiry: expiry})
	if m.signedVideoErr != nil {
		return "", time.Time{}, m.signedVideoErr
	}
	url := m.signedVideoURL
	if url == "" {
		url = "https://cdn.test/api/v1/uploads/videos/" + videoID + "/" + filename + "?op=video-upload&exp=0&keyId=v1&sig=fake"
	}
	return url, time.Now().Add(expiry), nil
}

func (m *mockCDNClient) UploadThumbnail(_ context.Context, videoID, ext, contentType string, body io.Reader) (string, error) {
	buf, _ := io.ReadAll(body)
	m.uploadedThumbnails = append(m.uploadedThumbnails, uploadedThumbnail{
		videoID:     videoID,
		ext:         ext,
		contentType: contentType,
		body:        buf,
	})
	if m.uploadThumbnailErr != nil {
		return "", m.uploadThumbnailErr
	}
	if m.uploadThumbnailObjectKey != "" {
		return m.uploadThumbnailObjectKey, nil
	}
	return "videos/" + videoID + "/thumbnail" + ext, nil
}

func (m *mockCDNClient) ThumbnailURL(videoID string) string {
	return "https://cdn.test/api/v1/thumbnails/" + videoID
}

// --- Test Helpers ---

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestService(repo *mockVideoRepo, dispatcher *mockDispatcher, opts ...Option) *Service {
	return newTestServiceWithStorage(repo, dispatcher, &stubStorage{}, opts...)
}

func newTestServiceWithStorage(repo *mockVideoRepo, dispatcher *mockDispatcher, store *stubStorage, opts ...Option) *Service {
	if repo == nil {
		repo = &mockVideoRepo{}
	}
	if dispatcher == nil {
		dispatcher = &mockDispatcher{}
	}
	if store == nil {
		store = &stubStorage{}
	}
	return NewService(repo, stubGenreRepo{}, stubSubRepo{}, stubPlanRepo{}, store, dispatcher, newTestLogger(), opts...)
}
