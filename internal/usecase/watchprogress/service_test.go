package watchprogress

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// --- Mock VideoRepository (only the methods this service uses) ---

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
func (m *mockVideoRepo) ListByUploader(context.Context, uuid.UUID, *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}
func (m *mockVideoRepo) SetSeriesEpisode(context.Context, uuid.UUID, uuid.UUID, int, int) error {
	return nil
}
func (m *mockVideoRepo) RemoveFromSeries(context.Context, uuid.UUID) error { return nil }
func (m *mockVideoRepo) ListBySeries(context.Context, uuid.UUID, *filter.FilterSet) ([]entity.Video, int64, error) {
	return nil, 0, nil
}

// --- Mock WatchProgressRepository ---

type mockProgressRepo struct {
	upsertFn         func(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int64) error
	getFn            func(ctx context.Context, userID, videoID uuid.UUID) (*entity.WatchProgress, error)
	deleteFn         func(ctx context.Context, userID, videoID uuid.UUID) error
	listInProgressFn func(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.VideoWithProgress, int64, error)
}

func (m *mockProgressRepo) Upsert(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int64) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, userID, videoID, positionSeconds)
	}
	return nil
}

func (m *mockProgressRepo) Get(ctx context.Context, userID, videoID uuid.UUID) (*entity.WatchProgress, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID, videoID)
	}
	return nil, nil
}

func (m *mockProgressRepo) Delete(ctx context.Context, userID, videoID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, videoID)
	}
	return nil
}

func (m *mockProgressRepo) ListInProgress(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.VideoWithProgress, int64, error) {
	if m.listInProgressFn != nil {
		return m.listInProgressFn(ctx, userID, page)
	}
	return nil, 0, nil
}

// --- Helpers ---

func newTestService(pr *mockProgressRepo, vr *mockVideoRepo) *Service {
	if pr == nil {
		pr = &mockProgressRepo{}
	}
	if vr == nil {
		vr = &mockVideoRepo{}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(pr, vr, log)
}

func newPublishedVideo(durationSeconds int64) *entity.Video {
	slug, _ := valueobject.NewSlug("clip")
	v := entity.NewVideo("Clip", slug, "", "", "G", "videos/x.mp4", "video/mp4", 0, uuid.New())
	v.Duration = valueobject.NewDuration(durationSeconds)
	return v
}

// --- Tests ---

func TestService_Upsert(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("happy path saves position", func(t *testing.T) {
		v := newPublishedVideo(600)
		var savedPos int64 = -1
		svc := newTestService(
			&mockProgressRepo{upsertFn: func(_ context.Context, u, vid uuid.UUID, pos int64) error {
				assert.Equal(t, userID, u)
				assert.Equal(t, v.ID, vid)
				savedPos = pos
				return nil
			}},
			&mockVideoRepo{getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return v, nil
			}},
		)

		appErr := svc.Upsert(ctx, userID, v.ID, 120)

		require.Nil(t, appErr)
		assert.Equal(t, int64(120), savedPos)
	})

	t.Run("negative position is rejected", func(t *testing.T) {
		svc := newTestService(nil, nil)

		appErr := svc.Upsert(ctx, userID, uuid.New(), -1)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeBadRequest, appErr.Code)
	})

	t.Run("position past duration is clamped", func(t *testing.T) {
		v := newPublishedVideo(600)
		var savedPos int64 = -1
		svc := newTestService(
			&mockProgressRepo{upsertFn: func(_ context.Context, _, _ uuid.UUID, pos int64) error {
				savedPos = pos
				return nil
			}},
			&mockVideoRepo{getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return v, nil
			}},
		)

		appErr := svc.Upsert(ctx, userID, v.ID, 999)

		require.Nil(t, appErr)
		assert.Equal(t, int64(600), savedPos)
	})

	t.Run("video not found returns 404", func(t *testing.T) {
		svc := newTestService(nil, &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return nil, nil },
		})

		appErr := svc.Upsert(ctx, userID, uuid.New(), 30)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotFound, appErr.Code)
	})

	t.Run("video repo error returns internal", func(t *testing.T) {
		svc := newTestService(nil, &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return nil, errors.New("boom")
			},
		})

		appErr := svc.Upsert(ctx, userID, uuid.New(), 30)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})

	t.Run("upsert repo error returns internal", func(t *testing.T) {
		v := newPublishedVideo(600)
		svc := newTestService(
			&mockProgressRepo{upsertFn: func(_ context.Context, _, _ uuid.UUID, _ int64) error {
				return errors.New("boom")
			}},
			&mockVideoRepo{getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return v, nil
			}},
		)

		appErr := svc.Upsert(ctx, userID, v.ID, 30)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_Get(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("returns progress and duration", func(t *testing.T) {
		v := newPublishedVideo(600)
		now := time.Now().UTC()
		svc := newTestService(
			&mockProgressRepo{getFn: func(_ context.Context, _, _ uuid.UUID) (*entity.WatchProgress, error) {
				return &entity.WatchProgress{
					UserID:          userID,
					VideoID:         v.ID,
					PositionSeconds: 120,
					UpdatedAt:       now,
				}, nil
			}},
			&mockVideoRepo{getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return v, nil
			}},
		)

		progress, duration, appErr := svc.Get(ctx, userID, v.ID)

		require.Nil(t, appErr)
		require.NotNil(t, progress)
		assert.Equal(t, int64(120), progress.PositionSeconds)
		assert.Equal(t, int64(600), duration)
	})

	t.Run("missing progress returns 404 with watch progress code", func(t *testing.T) {
		v := newPublishedVideo(600)
		svc := newTestService(
			&mockProgressRepo{getFn: func(_ context.Context, _, _ uuid.UUID) (*entity.WatchProgress, error) {
				return nil, nil
			}},
			&mockVideoRepo{getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return v, nil
			}},
		)

		_, _, appErr := svc.Get(ctx, userID, v.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeWatchProgressNotFound, appErr.Code)
	})

	t.Run("missing video returns 404 with video code", func(t *testing.T) {
		svc := newTestService(nil, &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return nil, nil },
		})

		_, _, appErr := svc.Get(ctx, userID, uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotFound, appErr.Code)
	})
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	videoID := uuid.New()

	t.Run("happy path", func(t *testing.T) {
		svc := newTestService(
			&mockProgressRepo{deleteFn: func(_ context.Context, _, _ uuid.UUID) error { return nil }},
			nil,
		)
		assert.Nil(t, svc.Delete(ctx, userID, videoID))
	})

	t.Run("missing entry returns 404", func(t *testing.T) {
		svc := newTestService(
			&mockProgressRepo{deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
				return domain.ErrWatchProgressNotFound
			}},
			nil,
		)

		appErr := svc.Delete(ctx, userID, videoID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeWatchProgressNotFound, appErr.Code)
	})

	t.Run("repo error returns internal", func(t *testing.T) {
		svc := newTestService(
			&mockProgressRepo{deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
				return errors.New("boom")
			}},
			nil,
		)

		appErr := svc.Delete(ctx, userID, videoID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_ListInProgress(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	page := valueobject.NewPagination(1, 10)

	t.Run("delegates to repo", func(t *testing.T) {
		v := newPublishedVideo(600)
		svc := newTestService(
			&mockProgressRepo{listInProgressFn: func(_ context.Context, u uuid.UUID, p valueobject.Pagination) ([]entity.VideoWithProgress, int64, error) {
				assert.Equal(t, userID, u)
				assert.Equal(t, page, p)
				return []entity.VideoWithProgress{
					{
						Video: *v,
						Progress: entity.WatchProgress{
							UserID:          userID,
							VideoID:         v.ID,
							PositionSeconds: 120,
							UpdatedAt:       time.Now().UTC(),
						},
					},
				}, 1, nil
			}},
			nil,
		)

		items, total, appErr := svc.ListInProgress(ctx, userID, page)

		require.Nil(t, appErr)
		assert.Equal(t, int64(1), total)
		require.Len(t, items, 1)
		assert.Equal(t, int64(120), items[0].Progress.PositionSeconds)
	})

	t.Run("repo error returns internal", func(t *testing.T) {
		svc := newTestService(
			&mockProgressRepo{listInProgressFn: func(_ context.Context, _ uuid.UUID, _ valueobject.Pagination) ([]entity.VideoWithProgress, int64, error) {
				return nil, 0, errors.New("boom")
			}},
			nil,
		)

		_, _, appErr := svc.ListInProgress(ctx, userID, page)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
