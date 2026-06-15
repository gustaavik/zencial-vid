package analytics

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestRecordPlayback(t *testing.T) {
	videoID := uuid.New()
	userID := uuid.New()
	sessionID := uuid.New()

	baseInput := func() *RecordPlaybackInput {
		return &RecordPlaybackInput{
			SessionID:       sessionID,
			VideoID:         videoID,
			UserID:          userID,
			Source:          "home",
			Platform:        "ios",
			CountryCode:     "DK",
			PositionSeconds: 120,
			WatchedSeconds:  90,
			WatchedBuckets:  []int{0, 1, 2},
			Completed:       false,
		}
	}

	tests := []struct {
		name       string
		input      func() *RecordPlaybackInput
		video      *entity.Video
		videoErr   error
		upsertErr  error
		wantCode   string
		checkHB    func(t *testing.T, hb *repository.PlaybackHeartbeat)
		wantNoCall bool
	}{
		{
			name:  "success passes normalized heartbeat",
			input: baseInput,
			video: testVideo(videoID, userID, 600),
			checkHB: func(t *testing.T, hb *repository.PlaybackHeartbeat) {
				if hb.SessionID != sessionID || hb.VideoID != videoID {
					t.Errorf("heartbeat ids = (%s,%s), want (%s,%s)", hb.SessionID, hb.VideoID, sessionID, videoID)
				}
				if hb.UserID == nil || *hb.UserID != userID {
					t.Errorf("heartbeat user = %v, want %s", hb.UserID, userID)
				}
				if hb.Source != entity.PlaybackSourceHome {
					t.Errorf("source = %q, want home", hb.Source)
				}
				if hb.Platform != entity.PlaybackPlatformIOS {
					t.Errorf("platform = %q, want ios", hb.Platform)
				}
				if hb.ViewThresholdSeconds != 30 {
					t.Errorf("threshold = %d, want 30", hb.ViewThresholdSeconds)
				}
			},
		},
		{
			name: "unknown source and platform are coerced to other",
			input: func() *RecordPlaybackInput {
				in := baseInput()
				in.Source = "tiktok"
				in.Platform = "windows"
				return in
			},
			video: testVideo(videoID, userID, 600),
			checkHB: func(t *testing.T, hb *repository.PlaybackHeartbeat) {
				if hb.Source != entity.PlaybackSourceOther {
					t.Errorf("source = %q, want other", hb.Source)
				}
				if hb.Platform != entity.PlaybackPlatformOther {
					t.Errorf("platform = %q, want other", hb.Platform)
				}
			},
		},
		{
			name: "position clamped to duration and watched capped at 3x duration",
			input: func() *RecordPlaybackInput {
				in := baseInput()
				in.PositionSeconds = 10_000
				in.WatchedSeconds = 10_000
				return in
			},
			video: testVideo(videoID, userID, 600),
			checkHB: func(t *testing.T, hb *repository.PlaybackHeartbeat) {
				if hb.PositionSeconds != 600 {
					t.Errorf("position = %d, want 600", hb.PositionSeconds)
				}
				if hb.WatchedSeconds != 1800 {
					t.Errorf("watched = %d, want 1800", hb.WatchedSeconds)
				}
			},
		},
		{
			name: "short video threshold uses half duration",
			input: func() *RecordPlaybackInput {
				in := baseInput()
				in.PositionSeconds = 5
				in.WatchedSeconds = 5
				return in
			},
			video: testVideo(videoID, userID, 20),
			checkHB: func(t *testing.T, hb *repository.PlaybackHeartbeat) {
				if hb.ViewThresholdSeconds != 10 {
					t.Errorf("threshold = %d, want 10", hb.ViewThresholdSeconds)
				}
			},
		},
		{
			name: "negative position rejected",
			input: func() *RecordPlaybackInput {
				in := baseInput()
				in.PositionSeconds = -1
				return in
			},
			wantCode:   apperror.CodeBadRequest,
			wantNoCall: true,
		},
		{
			name:       "video not found",
			input:      baseInput,
			video:      nil,
			wantCode:   apperror.CodeVideoNotFound,
			wantNoCall: true,
		},
		{
			name:       "video lookup error",
			input:      baseInput,
			videoErr:   errors.New("db down"),
			wantCode:   apperror.CodeInternalError,
			wantNoCall: true,
		},
		{
			name:      "session conflict is swallowed",
			input:     baseInput,
			video:     testVideo(videoID, userID, 600),
			upsertErr: domain.ErrPlaybackSessionConflict,
		},
		{
			name:      "repo error surfaces as internal",
			input:     baseInput,
			video:     testVideo(videoID, userID, 600),
			upsertErr: errors.New("db down"),
			wantCode:  apperror.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured *repository.PlaybackHeartbeat
			playbackRepo := &mockPlaybackRepo{
				upsertHeartbeatFn: func(_ context.Context, hb *repository.PlaybackHeartbeat) error {
					captured = hb
					return tt.upsertErr
				},
			}
			videoRepo := &mockVideoRepo{
				getByIDFn: func(context.Context, uuid.UUID) (*entity.Video, error) {
					return tt.video, tt.videoErr
				},
			}
			svc := newTestService(nil, playbackRepo, videoRepo)

			appErr := svc.RecordPlayback(context.Background(), tt.input())

			if tt.wantCode != "" {
				if appErr == nil || appErr.Code != tt.wantCode {
					t.Fatalf("error = %v, want code %s", appErr, tt.wantCode)
				}
			} else if appErr != nil {
				t.Fatalf("unexpected error: %v", appErr)
			}

			if tt.wantNoCall && captured != nil {
				t.Fatal("expected no heartbeat upsert")
			}
			if tt.checkHB != nil {
				if captured == nil {
					t.Fatal("expected heartbeat upsert")
				}
				tt.checkHB(t, captured)
			}
		})
	}
}
