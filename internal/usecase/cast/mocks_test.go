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
	createFn      func(ctx context.Context, cast *entity.Cast) error
	getByIDFn     func(ctx context.Context, id uuid.UUID) (*entity.Cast, error)
	listByVideoFn func(ctx context.Context, videoID uuid.UUID) ([]entity.Cast, error)
	updateFn      func(ctx context.Context, cast *entity.Cast) error
	deleteFn      func(ctx context.Context, id uuid.UUID) error
}

func (m *mockCastRepo) Create(ctx context.Context, cast *entity.Cast) error {
	return m.createFn(ctx, cast)
}
func (m *mockCastRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Cast, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockCastRepo) ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.Cast, error) {
	return m.listByVideoFn(ctx, videoID)
}
func (m *mockCastRepo) Update(ctx context.Context, cast *entity.Cast) error {
	return m.updateFn(ctx, cast)
}
func (m *mockCastRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
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
func newTestService(castRepo *mockCastRepo, videoRepo *mockVideoRepo, storageSvc storage.StorageService) *Service {
	return NewService(castRepo, videoRepo, slog.Default(), storageSvc)
}

// newActiveVideo returns a minimal Video entity owned by ownerID.
func newActiveVideo(videoID, ownerID uuid.UUID) *entity.Video {
	return &entity.Video{
		ID:         videoID,
		UploadedBy: ownerID,
	}
}

// newCastMember returns a minimal Cast entity.
func newCastMember(castID, videoID uuid.UUID) *entity.Cast {
	now := time.Now().UTC()
	return &entity.Cast{
		ID:        castID,
		VideoID:   videoID,
		Name:      "Jane Doe",
		Role:      "actor",
		SortOrder: 0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
