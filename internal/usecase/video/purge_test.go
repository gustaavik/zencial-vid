package video

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// objectPresent is a non-nil ObjectInfo used to signal "file exists in S3".
var objectPresent = &storage.ObjectInfo{Size: 100, ContentType: "video/mp4"}

func TestService_PurgeOrphans_Phase1_DBOrphans(t *testing.T) {
	ctx := context.Background()

	t.Run("no videos returns empty output", func(t *testing.T) {
		svc := newTestService(nil, nil)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{})

		require.Nil(t, appErr)
		assert.Empty(t, out.DBOrphans)
		assert.Empty(t, out.S3Orphans)
	})

	t.Run("all videos exist in S3 returns no orphans", func(t *testing.T) {
		id := uuid.New()
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: "videos/a.mp4"},
					}, nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return objectPresent, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{})

		require.Nil(t, appErr)
		assert.Empty(t, out.DBOrphans)
	})

	t.Run("video missing from S3 is reported and deleted", func(t *testing.T) {
		id := uuid.New()
		var deletedID uuid.UUID
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: "videos/missing.mp4"},
					}, nil
				},
				deleteFn: func(_ context.Context, vid uuid.UUID) error {
					deletedID = vid
					return nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return nil, nil // file not found
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{DryRun: false})

		require.Nil(t, appErr)
		require.Len(t, out.DBOrphans, 1)
		assert.Equal(t, id, out.DBOrphans[0])
		assert.Equal(t, id, deletedID)
	})

	t.Run("multiple orphans are all deleted", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		var deletedIDs []uuid.UUID
		infos := make([]repository.VideoStorageInfo, len(ids))
		for i, id := range ids {
			infos[i] = repository.VideoStorageInfo{ID: id, StorageKey: "videos/x.mp4"}
		}
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return infos, nil
				},
				deleteFn: func(_ context.Context, id uuid.UUID) error {
					deletedIDs = append(deletedIDs, id)
					return nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return nil, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{})

		require.Nil(t, appErr)
		assert.Len(t, out.DBOrphans, 3)
		assert.ElementsMatch(t, ids, deletedIDs)
	})

	t.Run("dry_run reports orphans without deleting", func(t *testing.T) {
		id := uuid.New()
		deleteCallCount := 0
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: "videos/missing.mp4"},
					}, nil
				},
				deleteFn: func(_ context.Context, _ uuid.UUID) error {
					deleteCallCount++
					return nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return nil, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{DryRun: true})

		require.Nil(t, appErr)
		assert.Len(t, out.DBOrphans, 1)
		assert.Equal(t, 0, deleteCallCount, "repo Delete must not be called in dry-run mode")
	})

	t.Run("mixed: some present some missing", func(t *testing.T) {
		present := uuid.New()
		missing := uuid.New()
		var deletedID uuid.UUID
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: present, StorageKey: "videos/exists.mp4"},
						{ID: missing, StorageKey: "videos/gone.mp4"},
					}, nil
				},
				deleteFn: func(_ context.Context, id uuid.UUID) error {
					deletedID = id
					return nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, key string) (*storage.ObjectInfo, error) {
					if key == "videos/exists.mp4" {
						return objectPresent, nil
					}
					return nil, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{})

		require.Nil(t, appErr)
		require.Len(t, out.DBOrphans, 1)
		assert.Equal(t, missing, out.DBOrphans[0])
		assert.Equal(t, missing, deletedID)
	})
}

func TestService_PurgeOrphans_Phase1_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("ListAllStorageKeys error returns internal", func(t *testing.T) {
		svc := newTestService(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return nil, errors.New("db unavailable")
				},
			},
			nil,
		)

		_, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})

	t.Run("Stat error returns internal", func(t *testing.T) {
		id := uuid.New()
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{{ID: id, StorageKey: "videos/x.mp4"}}, nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return nil, errors.New("s3 unreachable")
				},
			},
		)

		_, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})

	t.Run("Delete repo error returns internal", func(t *testing.T) {
		id := uuid.New()
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{{ID: id, StorageKey: "videos/x.mp4"}}, nil
				},
				deleteFn: func(_ context.Context, _ uuid.UUID) error {
					return errors.New("db error")
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return nil, nil // orphan
				},
			},
		)

		_, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_PurgeOrphans_Phase2_S3Orphans(t *testing.T) {
	ctx := context.Background()

	t.Run("all S3 keys match DB storage keys, no S3 orphans", func(t *testing.T) {
		id := uuid.New()
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: "videos/a.mp4"},
					}, nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return objectPresent, nil
				},
				listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
					return []string{"videos/a.mp4"}, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true})

		require.Nil(t, appErr)
		assert.Empty(t, out.DBOrphans)
		assert.Empty(t, out.S3Orphans)
	})

	t.Run("S3 key not in DB is reported and deleted", func(t *testing.T) {
		id := uuid.New()
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: "videos/known.mp4"},
					}, nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return objectPresent, nil
				},
				listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
					return []string{"videos/known.mp4", "videos/orphan.mp4"}, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true})

		require.Nil(t, appErr)
		assert.Empty(t, out.DBOrphans)
		require.Len(t, out.S3Orphans, 1)
		assert.Equal(t, "videos/orphan.mp4", out.S3Orphans[0])
	})

	t.Run("thumbnail key prevents false S3 orphan", func(t *testing.T) {
		id := uuid.New()
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: "videos/v.mp4", ThumbnailKey: "videos/v/thumb.jpg"},
					}, nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return objectPresent, nil
				},
				listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
					return []string{"videos/v.mp4", "videos/v/thumb.jpg"}, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true})

		require.Nil(t, appErr)
		assert.Empty(t, out.S3Orphans)
	})

	t.Run("dry_run reports S3 orphans without deleting", func(t *testing.T) {
		id := uuid.New()
		store := &stubStorage{
			statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
				return objectPresent, nil
			},
			listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
				return []string{"videos/known.mp4", "videos/orphan.mp4"}, nil
			},
		}
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: "videos/known.mp4"},
					}, nil
				},
			},
			nil,
			store,
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true, DryRun: true})

		require.Nil(t, appErr)
		require.Len(t, out.S3Orphans, 1)
		assert.Equal(t, "videos/orphan.mp4", out.S3Orphans[0])
		assert.Empty(t, store.deletedKeys, "storage Delete must not be called in dry-run mode")
	})

	t.Run("S3 orphans not checked when IncludeS3Orphans is false", func(t *testing.T) {
		listCalled := false
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return nil, nil
				},
			},
			nil,
			&stubStorage{
				listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
					listCalled = true
					return nil, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: false})

		require.Nil(t, appErr)
		assert.Empty(t, out.S3Orphans)
		assert.False(t, listCalled, "ListObjects must not be called when IncludeS3Orphans is false")
	})

	t.Run("HLS files for a known video are not S3 orphans", func(t *testing.T) {
		id := uuid.New()
		hlsPrefix := fmt.Sprintf("videos/%s/hls/", id)
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: fmt.Sprintf("videos/%s/original.mp4", id)},
					}, nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return objectPresent, nil
				},
				listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
					return []string{
						fmt.Sprintf("videos/%s/original.mp4", id),
						hlsPrefix + "master.m3u8",
						hlsPrefix + "360p/playlist.m3u8",
						hlsPrefix + "360p/segment_000.ts",
					}, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true})

		require.Nil(t, appErr)
		assert.Empty(t, out.S3Orphans)
	})

	t.Run("HLS files for an unknown videoID are S3 orphans", func(t *testing.T) {
		unknownID := uuid.New()
		hlsKey := fmt.Sprintf("videos/%s/hls/master.m3u8", unknownID)
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return nil, nil
				},
			},
			nil,
			&stubStorage{
				listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
					return []string{hlsKey}, nil
				},
			},
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true})

		require.Nil(t, appErr)
		require.Len(t, out.S3Orphans, 1)
		assert.Equal(t, hlsKey, out.S3Orphans[0])
	})

	t.Run("ListObjects error returns internal", func(t *testing.T) {
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return nil, nil
				},
			},
			nil,
			&stubStorage{
				listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
					return nil, errors.New("s3 error")
				},
			},
		)

		_, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})

	t.Run("S3 Delete error returns internal", func(t *testing.T) {
		id := uuid.New()
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: id, StorageKey: "videos/known.mp4"},
					}, nil
				},
			},
			nil,
			&stubStorage{
				statFn: func(_ context.Context, _ string) (*storage.ObjectInfo, error) {
					return objectPresent, nil
				},
				listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
					return []string{"videos/known.mp4", "videos/orphan.mp4"}, nil
				},
				deleteFn: func(_ context.Context, _ string) error {
					return errors.New("delete failed")
				},
			},
		)

		_, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_PurgeOrphans_BothPhases(t *testing.T) {
	ctx := context.Background()

	t.Run("DB and S3 orphans detected and removed together", func(t *testing.T) {
		existsID := uuid.New()
		missingID := uuid.New()

		var deletedVideoID uuid.UUID
		store := &stubStorage{
			statFn: func(_ context.Context, key string) (*storage.ObjectInfo, error) {
				if key == "videos/exists.mp4" {
					return objectPresent, nil
				}
				return nil, nil // "videos/missing.mp4" → orphan
			},
			listObjectsFn: func(_ context.Context, _ string) ([]string, error) {
				// "videos/missing.mp4" is gone from S3 so won't appear here.
				// "videos/exists.mp4" is known. "videos/s3-only.mp4" is an S3 orphan.
				return []string{"videos/exists.mp4", "videos/s3-only.mp4"}, nil
			},
		}
		svc := newTestServiceWithStorage(
			&mockVideoRepo{
				listAllStorageKeysFn: func(_ context.Context) ([]repository.VideoStorageInfo, error) {
					return []repository.VideoStorageInfo{
						{ID: existsID, StorageKey: "videos/exists.mp4"},
						{ID: missingID, StorageKey: "videos/missing.mp4"},
					}, nil
				},
				deleteFn: func(_ context.Context, id uuid.UUID) error {
					deletedVideoID = id
					return nil
				},
			},
			nil,
			store,
		)

		out, appErr := svc.PurgeOrphans(ctx, PurgeOrphansInput{IncludeS3Orphans: true})

		require.Nil(t, appErr)
		require.Len(t, out.DBOrphans, 1)
		assert.Equal(t, missingID, out.DBOrphans[0])
		assert.Equal(t, missingID, deletedVideoID)
		require.Len(t, out.S3Orphans, 1)
		assert.Equal(t, "videos/s3-only.mp4", out.S3Orphans[0])
		assert.Contains(t, store.deletedKeys, "videos/s3-only.mp4")
	})
}
