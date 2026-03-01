package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStreamSession(t *testing.T) {
	userID := uuid.New()
	contentID := uuid.New()
	episodeID := uuid.New()

	t.Run("with episode", func(t *testing.T) {
		session := NewStreamSession(userID, contentID, &episodeID, "Chrome/120", "192.168.1.1")

		require.NotNil(t, session)
		assert.NotEqual(t, uuid.Nil, session.ID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, contentID, session.ContentID)
		require.NotNil(t, session.EpisodeID)
		assert.Equal(t, episodeID, *session.EpisodeID)
		assert.Equal(t, "Chrome/120", session.DeviceInfo)
		assert.Equal(t, "192.168.1.1", session.IPAddress)
		assert.False(t, session.StartedAt.IsZero())
		assert.False(t, session.LastActiveAt.IsZero())
		assert.Equal(t, session.StartedAt, session.LastActiveAt)
	})

	t.Run("without episode", func(t *testing.T) {
		session := NewStreamSession(userID, contentID, nil, "Safari/17", "10.0.0.1")

		require.NotNil(t, session)
		assert.Nil(t, session.EpisodeID)
		assert.Equal(t, "Safari/17", session.DeviceInfo)
	})
}

func TestPlaybackProgress_Percentage(t *testing.T) {
	tests := []struct {
		name     string
		position int64
		duration int64
		want     float64
	}{
		{"half watched", 50, 100, 50.0},
		{"fully watched", 100, 100, 100.0},
		{"not started", 0, 100, 0.0},
		{"quarter watched", 25, 100, 25.0},
		{"zero duration", 50, 0, 0.0},
		{"90 percent", 90, 100, 90.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlaybackProgress{
				Position: tt.position,
				Duration: tt.duration,
			}
			assert.InDelta(t, tt.want, p.Percentage(), 0.001)
		})
	}
}

func TestPlaybackProgress_MarkCompleted(t *testing.T) {
	tests := []struct {
		name          string
		position      int64
		duration      int64
		wantCompleted bool
	}{
		{"exactly 90% marks completed", 90, 100, true},
		{"above 90% marks completed", 95, 100, true},
		{"100% marks completed", 100, 100, true},
		{"below 90% stays incomplete", 89, 100, false},
		{"50% stays incomplete", 50, 100, false},
		{"0% stays incomplete", 0, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlaybackProgress{
				Position:  tt.position,
				Duration:  tt.duration,
				Completed: false,
			}
			p.MarkCompleted()
			assert.Equal(t, tt.wantCompleted, p.Completed)
		})
	}
}
