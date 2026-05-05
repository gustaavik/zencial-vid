package audit

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// --- Test doubles ---

type fakeRepo struct {
	mu      sync.Mutex
	written []*entity.AuditLog
	err     error
}

func (r *fakeRepo) Create(_ context.Context, log *entity.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	r.written = append(r.written, log)
	return nil
}

func (r *fakeRepo) List(_ context.Context, _ *filter.FilterSet) ([]entity.AuditLog, int64, error) {
	return nil, 0, nil
}

func (r *fakeRepo) all() []*entity.AuditLog {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*entity.AuditLog, len(r.written))
	copy(out, r.written)
	return out
}

type fakeDispatcher struct {
	allHandlers []func(event.Event) error
}

func (d *fakeDispatcher) Dispatch(e event.Event) error {
	for _, h := range d.allHandlers {
		_ = h(e)
	}
	return nil
}
func (d *fakeDispatcher) Subscribe(_ string, _ func(event.Event) error) {}
func (d *fakeDispatcher) SubscribeAll(h func(event.Event) error) {
	d.allHandlers = append(d.allHandlers, h)
}

// nonAuditableEvent simulates an event that does not implement AuditableEvent.
type nonAuditableEvent struct{}

func (nonAuditableEvent) EventName() string     { return "non.auditable" }
func (nonAuditableEvent) OccurredAt() time.Time { return time.Now() }

func newQuietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// --- Tests ---

func TestRegister_PersistsAuditableEvent(t *testing.T) {
	repo := &fakeRepo{}
	dispatcher := &fakeDispatcher{}
	Register(dispatcher, repo, newQuietLogger())

	actorID := uuid.New()
	videoID := uuid.New()
	now := time.Now().UTC()

	err := dispatcher.Dispatch(event.VideoUploaded{
		VideoID:    videoID,
		Title:      "My Video",
		UploadedBy: actorID,
		Timestamp:  now,
	})
	require.NoError(t, err)

	logs := repo.all()
	require.Len(t, logs, 1)
	got := logs[0]
	assert.Equal(t, "video.uploaded", got.EventName)
	assert.Equal(t, "video", got.EntityType)
	require.NotNil(t, got.ActorID)
	assert.Equal(t, actorID, *got.ActorID)
	require.NotNil(t, got.EntityID)
	assert.Equal(t, videoID, *got.EntityID)
	assert.Equal(t, "My Video", got.Metadata["title"])
	assert.Equal(t, now, got.OccurredAt)
}

func TestRegister_SystemEventHasNilActor(t *testing.T) {
	repo := &fakeRepo{}
	dispatcher := &fakeDispatcher{}
	Register(dispatcher, repo, newQuietLogger())

	require.NoError(t, dispatcher.Dispatch(event.VideoTranscodeFailed{
		VideoID:   uuid.New(),
		Reason:    "ffmpeg crashed",
		Timestamp: time.Now().UTC(),
	}))

	logs := repo.all()
	require.Len(t, logs, 1)
	assert.Nil(t, logs[0].ActorID, "CDN-callback events must record nil actor")
	assert.Equal(t, "ffmpeg crashed", logs[0].Metadata["reason"])
}

func TestRegister_NonAuditableEventSkipped(t *testing.T) {
	repo := &fakeRepo{}
	dispatcher := &fakeDispatcher{}
	Register(dispatcher, repo, newQuietLogger())

	require.NoError(t, dispatcher.Dispatch(nonAuditableEvent{}))
	assert.Empty(t, repo.all(), "events not implementing AuditableEvent must be skipped silently")
}

func TestRegister_RepoErrorDoesNotPropagate(t *testing.T) {
	repo := &fakeRepo{err: errors.New("db down")}
	dispatcher := &fakeDispatcher{}
	Register(dispatcher, repo, newQuietLogger())

	err := dispatcher.Dispatch(event.UserLoggedIn{
		UserID:    uuid.New(),
		Timestamp: time.Now().UTC(),
	})
	assert.NoError(t, err, "audit failure must not break the originating operation")
}

func TestRegister_GenreEventsEntityType(t *testing.T) {
	repo := &fakeRepo{}
	dispatcher := &fakeDispatcher{}
	Register(dispatcher, repo, newQuietLogger())

	actorID := uuid.New()
	require.NoError(t, dispatcher.Dispatch(event.GenreCreated{
		GenreID:   uuid.New(),
		ActorID:   &actorID,
		Name:      "Action",
		Timestamp: time.Now().UTC(),
	}))

	logs := repo.all()
	require.Len(t, logs, 1)
	assert.Equal(t, "genre", logs[0].EntityType)
	assert.Equal(t, "Action", logs[0].Metadata["name"])
	require.NotNil(t, logs[0].ActorID)
	assert.Equal(t, actorID, *logs[0].ActorID)
}
