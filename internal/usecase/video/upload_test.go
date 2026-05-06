package video

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_InitiateUpload(t *testing.T) {
	ctx := context.Background()

	t.Run("returns presigned URL with content-typed extension", func(t *testing.T) {
		store := &stubStorage{
			presignPutFn: func(_ context.Context, key, ct string, _ time.Duration) (string, error) {
				assert.Equal(t, "video/mp4", ct)
				assert.True(t, strings.HasSuffix(key, ".mp4"), "key should end with .mp4: %s", key)
				assert.True(t, strings.HasPrefix(key, "videos/"), "key should start with videos/: %s", key)
				return "https://s3.example/" + key + "?sig=stub", nil
			},
		}
		svc := newTestServiceWithStorage(nil, nil, store)

		out, appErr := svc.InitiateUpload(ctx, &InitiateUploadInput{
			FileName:    "my-clip.mp4",
			ContentType: "video/mp4",
		})

		require.Nil(t, appErr)
		require.NotNil(t, out)
		assert.Equal(t, store.presignPutCalls[0].Key, out.ObjectKey)
		assert.Equal(t, "https://s3.example/"+out.ObjectKey+"?sig=stub", out.UploadURL)
		assert.WithinDuration(t, time.Now().UTC().Add(PresignedUploadExpiry), out.ExpiresAt, time.Minute)

		// Object key encodes a parseable video UUID.
		id, err := videoIDFromObjectKey(out.ObjectKey)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
	})

	t.Run("falls back to filename extension when content type unknown", func(t *testing.T) {
		store := &stubStorage{}
		svc := newTestServiceWithStorage(nil, nil, store)

		out, appErr := svc.InitiateUpload(ctx, &InitiateUploadInput{
			FileName:    "raw.mkv",
			ContentType: "application/octet-stream",
		})

		require.Nil(t, appErr)
		assert.True(t, strings.HasSuffix(out.ObjectKey, ".mkv"), "key %s", out.ObjectKey)
	})

	t.Run("defaults to .mp4 when neither content type nor filename give an extension", func(t *testing.T) {
		store := &stubStorage{}
		svc := newTestServiceWithStorage(nil, nil, store)

		out, appErr := svc.InitiateUpload(ctx, &InitiateUploadInput{
			FileName:    "no-extension",
			ContentType: "application/octet-stream",
		})

		require.Nil(t, appErr)
		assert.True(t, strings.HasSuffix(out.ObjectKey, ".mp4"), "key %s", out.ObjectKey)
	})

	t.Run("storage error returns Internal apperror", func(t *testing.T) {
		store := &stubStorage{
			presignPutFn: func(context.Context, string, string, time.Duration) (string, error) {
				return "", errors.New("network down")
			},
		}
		svc := newTestServiceWithStorage(nil, nil, store)

		out, appErr := svc.InitiateUpload(ctx, &InitiateUploadInput{FileName: "x.mp4", ContentType: "video/mp4"})

		assert.Nil(t, out)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_CompleteUpload(t *testing.T) {
	ctx := context.Background()
	uploader := uuid.New()
	videoID := uuid.New()
	objectKey := "videos/" + videoID.String() + "/original.mp4"

	t.Run("happy path persists video and dispatches VideoUploaded", func(t *testing.T) {
		var created *entity.Video
		repo := &mockVideoRepo{
			createFn: func(_ context.Context, v *entity.Video) error {
				created = v
				return nil
			},
		}
		store := &stubStorage{
			statFn: func(_ context.Context, key string) (*storage.ObjectInfo, error) {
				assert.Equal(t, objectKey, key)
				return &storage.ObjectInfo{Size: 12345, ContentType: "video/mp4"}, nil
			},
		}
		dispatcher := &mockDispatcher{}
		svc := newTestServiceWithStorage(repo, dispatcher, store)

		out, appErr := svc.CompleteUpload(ctx, &CompleteUploadInput{
			ObjectKey:       objectKey,
			Title:           "My Video",
			UploadedBy:      uploader,
			DurationSeconds: 90,
		})

		require.Nil(t, appErr)
		require.NotNil(t, out)
		require.NotNil(t, created)
		assert.Equal(t, videoID, created.ID, "video ID is derived from the object key")
		assert.Equal(t, objectKey, created.StorageKey)
		assert.Equal(t, int64(12345), created.FileSize, "file size from storage Stat overrides any client claim")
		assert.Equal(t, uploader, created.UploadedBy)

		require.Len(t, dispatcher.dispatched, 1)
		evt, ok := dispatcher.dispatched[0].(event.VideoUploaded)
		require.True(t, ok)
		assert.Equal(t, videoID, evt.VideoID)
	})

	t.Run("missing object returns 400 — client never finished the PUT", func(t *testing.T) {
		store := &stubStorage{
			statFn: func(context.Context, string) (*storage.ObjectInfo, error) { return nil, nil },
		}
		repo := &mockVideoRepo{
			createFn: func(context.Context, *entity.Video) error {
				t.Fatal("Create should not be called when object is missing")
				return nil
			},
		}
		svc := newTestServiceWithStorage(repo, nil, store)

		out, appErr := svc.CompleteUpload(ctx, &CompleteUploadInput{
			ObjectKey:  objectKey,
			Title:      "X",
			UploadedBy: uploader,
		})

		assert.Nil(t, out)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeValidationFailed, appErr.Code)
	})

	t.Run("invalid object_key returns 400 without touching storage", func(t *testing.T) {
		store := &stubStorage{
			statFn: func(context.Context, string) (*storage.ObjectInfo, error) {
				t.Fatal("Stat should not be called for malformed key")
				return nil, nil
			},
		}
		svc := newTestServiceWithStorage(nil, nil, store)

		out, appErr := svc.CompleteUpload(ctx, &CompleteUploadInput{
			ObjectKey:  "not-a-valid-key",
			Title:      "X",
			UploadedBy: uploader,
		})

		assert.Nil(t, out)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeValidationFailed, appErr.Code)
	})

	t.Run("DB failure cleans up the orphaned object", func(t *testing.T) {
		repo := &mockVideoRepo{
			createFn: func(context.Context, *entity.Video) error { return errors.New("db down") },
		}
		store := &stubStorage{
			statFn: func(context.Context, string) (*storage.ObjectInfo, error) {
				return &storage.ObjectInfo{Size: 1, ContentType: "video/mp4"}, nil
			},
		}
		svc := newTestServiceWithStorage(repo, nil, store)

		out, appErr := svc.CompleteUpload(ctx, &CompleteUploadInput{
			ObjectKey:  objectKey,
			Title:      "X",
			UploadedBy: uploader,
		})

		assert.Nil(t, out)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
		assert.Contains(t, store.deletedKeys, objectKey, "orphaned object should be cleaned up on DB failure")
	})
}
