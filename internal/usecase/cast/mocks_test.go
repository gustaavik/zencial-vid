package cast

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// mockCastRepo is a closure-field mock for repository.CastRepository.
type mockCastRepo struct {
	createFn             func(ctx context.Context, cast *entity.Cast) error
	getByIDFn            func(ctx context.Context, id uuid.UUID) (*entity.Cast, error)
	getByNameFn          func(ctx context.Context, name string) (*entity.Cast, error)
	findOrCreateFn       func(ctx context.Context, name string) (*entity.Cast, error)
	updateFn             func(ctx context.Context, cast *entity.Cast) error
	deleteFn             func(ctx context.Context, id uuid.UUID) error
	hasVideoWithCallerFn func(ctx context.Context, castID, callerID uuid.UUID) (bool, error)
}

func (m *mockCastRepo) Create(ctx context.Context, cast *entity.Cast) error {
	return m.createFn(ctx, cast)
}
func (m *mockCastRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Cast, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockCastRepo) GetByName(ctx context.Context, name string) (*entity.Cast, error) {
	return m.getByNameFn(ctx, name)
}
func (m *mockCastRepo) FindOrCreate(ctx context.Context, name string) (*entity.Cast, error) {
	return m.findOrCreateFn(ctx, name)
}
func (m *mockCastRepo) Update(ctx context.Context, cast *entity.Cast) error {
	return m.updateFn(ctx, cast)
}
func (m *mockCastRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
}
func (m *mockCastRepo) HasVideoWithCaller(ctx context.Context, castID, callerID uuid.UUID) (bool, error) {
	return m.hasVideoWithCallerFn(ctx, castID, callerID)
}

// mockVideoCastRepo is a closure-field mock for repository.VideoCastRepository.
type mockVideoCastRepo struct {
	createFn               func(ctx context.Context, vc *entity.VideoCast) error
	getByVideoAndCastFn    func(ctx context.Context, videoID, castID uuid.UUID) (*entity.VideoCast, error)
	listByVideoFn          func(ctx context.Context, videoID uuid.UUID) ([]entity.VideoCast, error)
	updateFn               func(ctx context.Context, vc *entity.VideoCast) error
	deleteByVideoAndCastFn func(ctx context.Context, videoID, castID uuid.UUID) error
}

func (m *mockVideoCastRepo) Create(ctx context.Context, vc *entity.VideoCast) error {
	return m.createFn(ctx, vc)
}
func (m *mockVideoCastRepo) GetByVideoAndCast(ctx context.Context, videoID, castID uuid.UUID) (*entity.VideoCast, error) {
	return m.getByVideoAndCastFn(ctx, videoID, castID)
}
func (m *mockVideoCastRepo) ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.VideoCast, error) {
	return m.listByVideoFn(ctx, videoID)
}
func (m *mockVideoCastRepo) Update(ctx context.Context, vc *entity.VideoCast) error {
	return m.updateFn(ctx, vc)
}
func (m *mockVideoCastRepo) DeleteByVideoAndCast(ctx context.Context, videoID, castID uuid.UUID) error {
	return m.deleteByVideoAndCastFn(ctx, videoID, castID)
}

// mockVideoRepo is a closure-field mock for repository.VideoRepository (only GetByID is used by cast).
type mockVideoRepo struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*entity.Video, error)
}

func (m *mockVideoRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockVideoRepo) Create(_ context.Context, _ *entity.Video) error { return nil }
func (m *mockVideoRepo) GetBySlug(_ context.Context, _ valueobject.Slug) (*entity.Video, error) {
	return nil, nil
}
func (m *mockVideoRepo) Update(_ context.Context, _ *entity.Video) error { return nil }
func (m *mockVideoRepo) Delete(_ context.Context, _ uuid.UUID) error     { return nil }
func (m *mockVideoRepo) List(_ context.Context, _ *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}
func (m *mockVideoRepo) ListPublished(_ context.Context, _ *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}
func (m *mockVideoRepo) ListByUploader(_ context.Context, _ uuid.UUID, _ *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}
func (m *mockVideoRepo) ExistsBySlug(_ context.Context, _ valueobject.Slug) (bool, error) {
	return false, nil
}
func (m *mockVideoRepo) SetGenres(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error { return nil }
func (m *mockVideoRepo) GetGenreIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (m *mockVideoRepo) ListAllStorageKeys(_ context.Context) ([]repository.VideoStorageInfo, error) {
	return nil, nil
}
func (m *mockVideoRepo) SetSeriesEpisode(_ context.Context, _, _ uuid.UUID, _, _ int) error {
	return nil
}
func (m *mockVideoRepo) RemoveFromSeries(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockVideoRepo) ListBySeries(_ context.Context, _ uuid.UUID, _ *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}

// mockStorageSvc is a closure-field mock for storage.StorageService.
type mockStorageSvc struct {
	uploadFn    func(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	deleteFn    func(ctx context.Context, key string) error
	publicURLFn func(key string) string
}

func (m *mockStorageSvc) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	return m.uploadFn(ctx, key, body, contentType)
}
func (m *mockStorageSvc) Delete(ctx context.Context, key string) error {
	return m.deleteFn(ctx, key)
}
func (m *mockStorageSvc) PublicURL(key string) string {
	return m.publicURLFn(key)
}
func (m *mockStorageSvc) Move(_ context.Context, _, _ string) error { return nil }
func (m *mockStorageSvc) PresignedGetURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}
func (m *mockStorageSvc) PresignedPutURL(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "", nil
}
func (m *mockStorageSvc) Stat(_ context.Context, _ string) (*storage.ObjectInfo, error) {
	return nil, nil
}
func (m *mockStorageSvc) ListObjects(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

// newTestService creates a Service for testing.
// storageSvc accepts a storage.StorageService interface so that a true nil
// interface (not a typed-nil *mockStorageSvc) can be passed for the
// "storage not configured" test case.
func newTestService(castRepo *mockCastRepo, videoCastRepo *mockVideoCastRepo, videoRepo *mockVideoRepo, storageSvc storage.StorageService) *Service {
	return NewService(castRepo, videoCastRepo, videoRepo, slog.Default(), storageSvc)
}

// newCastMember returns a minimal Cast entity.
func newCastMember(castID uuid.UUID) *entity.Cast {
	now := time.Now().UTC()
	return &entity.Cast{
		ID:        castID,
		Name:      "Jane Doe",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
