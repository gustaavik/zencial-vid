package video

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func newProcessingVideo(t *testing.T) *entity.Video {
	t.Helper()
	slug, err := valueobject.NewSlug("clip")
	require.NoError(t, err)
	v := entity.NewVideo("Clip", slug, "", "", "G", "videos/x.mp4", "video/mp4", 0, uuid.New())
	v.Publish() // moves to processing
	return v
}

func TestService_MarkTranscodeComplete(t *testing.T) {
	ctx := context.Background()

	t.Run("processing → published dispatches VideoPublished", func(t *testing.T) {
		v := newProcessingVideo(t)
		var saved *entity.Video
		dispatcher := &mockDispatcher{}
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
			updateFn: func(_ context.Context, video *entity.Video) error {
				saved = video
				return nil
			},
		}, dispatcher)

		out, appErr := svc.MarkTranscodeComplete(ctx, v.ID)

		require.Nil(t, appErr)
		require.NotNil(t, out)
		assert.Equal(t, entity.VideoStatusPublished, out.Status)
		require.NotNil(t, saved, "repository.Update must be called")
		assert.Equal(t, entity.VideoStatusPublished, saved.Status)

		require.Len(t, dispatcher.dispatched, 1)
		_, ok := dispatcher.dispatched[0].(event.VideoPublished)
		assert.True(t, ok, "expected VideoPublished event")
	})

	t.Run("already published is idempotent — no event, no update", func(t *testing.T) {
		v := newProcessingVideo(t)
		v.MarkTranscoded()
		updateCalled := false
		dispatcher := &mockDispatcher{}
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
			updateFn: func(_ context.Context, _ *entity.Video) error {
				updateCalled = true
				return nil
			},
		}, dispatcher)

		out, appErr := svc.MarkTranscodeComplete(ctx, v.ID)

		require.Nil(t, appErr)
		require.NotNil(t, out)
		assert.False(t, updateCalled, "no DB write on idempotent retry")
		assert.Empty(t, dispatcher.dispatched, "no event re-fired")
	})

	t.Run("draft state rejected", func(t *testing.T) {
		v := newProcessingVideo(t)
		v.Status = entity.VideoStatusDraft
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
		}, nil)

		_, appErr := svc.MarkTranscodeComplete(ctx, v.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotTranscoding, appErr.Code)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return nil, nil },
		}, nil)

		_, appErr := svc.MarkTranscodeComplete(ctx, uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotFound, appErr.Code)
	})

	t.Run("repo update error returns internal error", func(t *testing.T) {
		v := newProcessingVideo(t)
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
			updateFn:  func(_ context.Context, _ *entity.Video) error { return errors.New("boom") },
		}, nil)

		_, appErr := svc.MarkTranscodeComplete(ctx, v.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_MarkTranscodeFailed(t *testing.T) {
	ctx := context.Background()

	t.Run("processing → failed records reason and dispatches event", func(t *testing.T) {
		v := newProcessingVideo(t)
		var saved *entity.Video
		dispatcher := &mockDispatcher{}
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
			updateFn: func(_ context.Context, video *entity.Video) error {
				saved = video
				return nil
			},
		}, dispatcher)

		out, appErr := svc.MarkTranscodeFailed(ctx, v.ID, "ffmpeg exit 1")

		require.Nil(t, appErr)
		assert.Equal(t, entity.VideoStatusFailed, out.Status)
		assert.Equal(t, "ffmpeg exit 1", out.TranscodeError)
		require.NotNil(t, saved)

		require.Len(t, dispatcher.dispatched, 1)
		failedEvt, ok := dispatcher.dispatched[0].(event.VideoTranscodeFailed)
		require.True(t, ok)
		assert.Equal(t, "ffmpeg exit 1", failedEvt.Reason)
	})

	t.Run("already failed updates reason but does not refire event", func(t *testing.T) {
		v := newProcessingVideo(t)
		v.MarkTranscodeFailed("first failure")
		dispatcher := &mockDispatcher{}
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
		}, dispatcher)

		out, appErr := svc.MarkTranscodeFailed(ctx, v.ID, "second failure")

		require.Nil(t, appErr)
		assert.Equal(t, "second failure", out.TranscodeError)
		assert.Empty(t, dispatcher.dispatched, "no event re-fired on idempotent retry")
	})

	t.Run("draft state rejected", func(t *testing.T) {
		v := newProcessingVideo(t)
		v.Status = entity.VideoStatusDraft
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
		}, nil)

		_, appErr := svc.MarkTranscodeFailed(ctx, v.ID, "anything")

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotTranscoding, appErr.Code)
	})
}

func TestService_Publish_MovesToProcessingAndTriggersCDN(t *testing.T) {
	ctx := context.Background()

	slug, err := valueobject.NewSlug("foo")
	require.NoError(t, err)
	v := entity.NewVideo("Foo", slug, "", "", "G", "videos/foo.mp4", "video/mp4", 0, uuid.New())

	cdn := &mockCDNClient{triggeredCh: make(chan string, 1)}
	dispatcher := &mockDispatcher{}
	svc := newTestService(&mockVideoRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
	}, dispatcher, WithCDN(cdn, "http://cdn"))

	out, appErr := svc.Publish(ctx, v.ID)
	require.Nil(t, appErr)
	assert.Equal(t, entity.VideoStatusProcessing, out.Status)
	assert.Empty(t, dispatcher.dispatched, "Publish must NOT dispatch VideoPublished — that happens on transcode complete")

	select {
	case got := <-cdn.triggeredCh:
		assert.Equal(t, v.ID.String(), got)
	case <-context.Background().Done():
	}
}
