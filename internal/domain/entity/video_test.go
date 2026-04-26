package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func newTestVideo(t *testing.T) *Video {
	t.Helper()
	slug, err := valueobject.NewSlug("test-video")
	require.NoError(t, err)
	return NewVideo(
		"Test Video",
		slug,
		"description",
		"creator",
		"G",
		"videos/abc.mp4",
		"video/mp4",
		1024,
		uuid.New(),
	)
}

func TestNewVideo_DefaultsToDraft(t *testing.T) {
	v := newTestVideo(t)
	assert.Equal(t, VideoStatusDraft, v.Status)
	assert.Empty(t, v.TranscodeError)
	assert.False(t, v.IsPlayable())
}

func TestVideoStatusTransitions(t *testing.T) {
	t.Run("Publish moves draft to processing", func(t *testing.T) {
		v := newTestVideo(t)
		v.Publish()
		assert.Equal(t, VideoStatusProcessing, v.Status)
		assert.Empty(t, v.TranscodeError)
		assert.False(t, v.IsPlayable(), "processing videos must not be playable")
	})

	t.Run("Publish clears prior transcode error when retrying from failed", func(t *testing.T) {
		v := newTestVideo(t)
		v.MarkTranscodeFailed("ffmpeg crashed")
		require.Equal(t, VideoStatusFailed, v.Status)
		require.Equal(t, "ffmpeg crashed", v.TranscodeError)

		v.Publish()
		assert.Equal(t, VideoStatusProcessing, v.Status)
		assert.Empty(t, v.TranscodeError, "publish must clear stale transcode error on retry")
	})

	t.Run("MarkTranscoded transitions processing to published", func(t *testing.T) {
		v := newTestVideo(t)
		v.Publish()
		v.MarkTranscoded()
		assert.Equal(t, VideoStatusPublished, v.Status)
		assert.Empty(t, v.TranscodeError)
		assert.True(t, v.IsPlayable())
	})

	t.Run("MarkTranscodeFailed records reason and sets failed status", func(t *testing.T) {
		v := newTestVideo(t)
		v.Publish()
		v.MarkTranscodeFailed("disk full")
		assert.Equal(t, VideoStatusFailed, v.Status)
		assert.Equal(t, "disk full", v.TranscodeError)
		assert.False(t, v.IsPlayable())
	})

	t.Run("Archive then Unarchive returns to draft", func(t *testing.T) {
		v := newTestVideo(t)
		v.Publish()
		v.MarkTranscoded()
		v.Archive()
		assert.Equal(t, VideoStatusArchived, v.Status)
		v.Unarchive()
		assert.Equal(t, VideoStatusDraft, v.Status)
	})
}

func TestIsPlayable_OnlyTrueWhenPublished(t *testing.T) {
	cases := []struct {
		status   VideoStatus
		playable bool
	}{
		{VideoStatusDraft, false},
		{VideoStatusProcessing, false},
		{VideoStatusPublished, true},
		{VideoStatusArchived, false},
		{VideoStatusFailed, false},
	}
	for _, tc := range cases {
		t.Run(string(tc.status), func(t *testing.T) {
			v := newTestVideo(t)
			v.Status = tc.status
			assert.Equal(t, tc.playable, v.IsPlayable())
		})
	}
}
