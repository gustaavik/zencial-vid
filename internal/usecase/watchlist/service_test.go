package watchlist

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

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

// --- Mock WatchlistRepository ---

type mockWatchlistRepo struct {
	addFn        func(ctx context.Context, userID, videoID uuid.UUID) error
	removeFn     func(ctx context.Context, userID, videoID uuid.UUID) error
	existsFn     func(ctx context.Context, userID, videoID uuid.UUID) (bool, error)
	listVideosFn func(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.Video, int64, error)
}

func (m *mockWatchlistRepo) Add(ctx context.Context, userID, videoID uuid.UUID) error {
	if m.addFn != nil {
		return m.addFn(ctx, userID, videoID)
	}
	return nil
}

func (m *mockWatchlistRepo) Remove(ctx context.Context, userID, videoID uuid.UUID) error {
	if m.removeFn != nil {
		return m.removeFn(ctx, userID, videoID)
	}
	return nil
}

func (m *mockWatchlistRepo) Exists(ctx context.Context, userID, videoID uuid.UUID) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, userID, videoID)
	}
	return false, nil
}

func (m *mockWatchlistRepo) ListVideos(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.Video, int64, error) {
	if m.listVideosFn != nil {
		return m.listVideosFn(ctx, userID, page)
	}
	return nil, 0, nil
}

// --- Helpers ---

func newTestService(wr *mockWatchlistRepo, vr *mockVideoRepo) *Service {
	if wr == nil {
		wr = &mockWatchlistRepo{}
	}
	if vr == nil {
		vr = &mockVideoRepo{}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(wr, vr, log)
}

func newPublishedVideo() *entity.Video {
	slug, _ := valueobject.NewSlug("clip")
	v := entity.NewVideo("Clip", slug, "", "", "G", "videos/x.mp4", "video/mp4", 0, uuid.New())
	return v
}

// --- Tests ---

func TestService_Add(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	v := newPublishedVideo()

	t.Run("happy path inserts entry", func(t *testing.T) {
		called := false
		svc := newTestService(
			&mockWatchlistRepo{addFn: func(_ context.Context, u, vid uuid.UUID) error {
				assert.Equal(t, userID, u)
				assert.Equal(t, v.ID, vid)
				called = true
				return nil
			}},
			&mockVideoRepo{getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return v, nil
			}},
		)

		appErr := svc.Add(ctx, userID, v.ID)

		require.Nil(t, appErr)
		assert.True(t, called, "Add must hit the repo")
	})

	t.Run("video not found returns 404", func(t *testing.T) {
		svc := newTestService(nil, &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return nil, nil },
		})

		appErr := svc.Add(ctx, userID, uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotFound, appErr.Code)
	})

	t.Run("idempotent: repo Add returning nil for duplicate is treated as success", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{addFn: func(_ context.Context, _, _ uuid.UUID) error { return nil }},
			&mockVideoRepo{getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return v, nil
			}},
		)
		assert.Nil(t, svc.Add(ctx, userID, v.ID))
		assert.Nil(t, svc.Add(ctx, userID, v.ID))
	})

	t.Run("video repo error returns internal", func(t *testing.T) {
		svc := newTestService(nil, &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) {
				return nil, errors.New("boom")
			},
		})

		appErr := svc.Add(ctx, userID, uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_Remove(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	videoID := uuid.New()

	t.Run("happy path", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{removeFn: func(_ context.Context, _, _ uuid.UUID) error { return nil }},
			nil,
		)
		assert.Nil(t, svc.Remove(ctx, userID, videoID))
	})

	t.Run("missing entry returns 404", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{removeFn: func(_ context.Context, _, _ uuid.UUID) error {
				return domain.ErrWatchlistEntryNotFound
			}},
			nil,
		)

		appErr := svc.Remove(ctx, userID, videoID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeWatchlistEntryNotFound, appErr.Code)
	})

	t.Run("repo error returns internal", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{removeFn: func(_ context.Context, _, _ uuid.UUID) error {
				return errors.New("boom")
			}},
			nil,
		)

		appErr := svc.Remove(ctx, userID, videoID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_IsInWatchlist(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	videoID := uuid.New()

	t.Run("true when present", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{existsFn: func(_ context.Context, _, _ uuid.UUID) (bool, error) {
				return true, nil
			}},
			nil,
		)

		got, appErr := svc.IsInWatchlist(ctx, userID, videoID)

		require.Nil(t, appErr)
		assert.True(t, got)
	})

	t.Run("false when absent", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{existsFn: func(_ context.Context, _, _ uuid.UUID) (bool, error) {
				return false, nil
			}},
			nil,
		)

		got, appErr := svc.IsInWatchlist(ctx, userID, videoID)

		require.Nil(t, appErr)
		assert.False(t, got)
	})

	t.Run("repo error returns internal", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{existsFn: func(_ context.Context, _, _ uuid.UUID) (bool, error) {
				return false, errors.New("boom")
			}},
			nil,
		)

		_, appErr := svc.IsInWatchlist(ctx, userID, videoID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_List(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	page := valueobject.NewPagination(1, 10)

	t.Run("returns videos and total", func(t *testing.T) {
		v := newPublishedVideo()
		svc := newTestService(
			&mockWatchlistRepo{listVideosFn: func(_ context.Context, u uuid.UUID, p valueobject.Pagination) ([]entity.Video, int64, error) {
				assert.Equal(t, userID, u)
				assert.Equal(t, page, p)
				return []entity.Video{*v}, 1, nil
			}},
			nil,
		)

		videos, total, appErr := svc.List(ctx, userID, page)

		require.Nil(t, appErr)
		assert.Equal(t, int64(1), total)
		require.Len(t, videos, 1)
		assert.Equal(t, v.ID, videos[0].ID)
	})

	t.Run("empty list returns zero", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{listVideosFn: func(_ context.Context, _ uuid.UUID, _ valueobject.Pagination) ([]entity.Video, int64, error) {
				return nil, 0, nil
			}},
			nil,
		)

		videos, total, appErr := svc.List(ctx, userID, page)

		require.Nil(t, appErr)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, videos)
	})

	t.Run("repo error returns internal", func(t *testing.T) {
		svc := newTestService(
			&mockWatchlistRepo{listVideosFn: func(_ context.Context, _ uuid.UUID, _ valueobject.Pagination) ([]entity.Video, int64, error) {
				return nil, 0, errors.New("boom")
			}},
			nil,
		)

		_, _, appErr := svc.List(ctx, userID, page)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
