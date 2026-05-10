package video

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
)

// CDNClient is the abstraction over the zencial-cdn service. The use cases
// depend only on this small interface; the concrete *cdn.Client implements
// the full surface but tests can substitute a fake.
type CDNClient interface {
	// TriggerTranscode kicks off async HLS transcoding for an uploaded video.
	TriggerTranscode(videoID string) error
	// SignVideoUploadURL mints a short-lived signed PUT URL the browser can
	// use to upload a video binary directly to the CDN.
	SignVideoUploadURL(videoID, filename string, expiry time.Duration) (string, time.Time, error)
	// UploadThumbnail proxies a thumbnail body from this API to the CDN over
	// the internal network and returns the resulting S3 object key.
	UploadThumbnail(ctx context.Context, videoID, ext, contentType string, body io.Reader) (string, error)
	// ThumbnailURL returns the public-facing GET URL for a video's thumbnail.
	ThumbnailURL(videoID string) string
}

// Service handles video use cases.
type Service struct {
	videoRepo  repository.VideoRepository
	genreRepo  repository.GenreRepository
	subRepo    repository.SubscriptionRepository
	planRepo   repository.PlanRepository
	storage    storage.StorageService
	dispatcher event.Dispatcher
	cdn        CDNClient
	cdnBaseURL string
	log        *slog.Logger
}

// NewService creates a new video Service.
func NewService(
	videoRepo repository.VideoRepository,
	genreRepo repository.GenreRepository,
	subRepo repository.SubscriptionRepository,
	planRepo repository.PlanRepository,
	storageSvc storage.StorageService,
	dispatcher event.Dispatcher,
	log *slog.Logger,
	opts ...Option,
) *Service {
	s := &Service{
		videoRepo:  videoRepo,
		genreRepo:  genreRepo,
		subRepo:    subRepo,
		planRepo:   planRepo,
		storage:    storageSvc,
		dispatcher: dispatcher,
		log:        log,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures optional Service dependencies.
type Option func(*Service)

// WithCDN configures the CDN client and base URL for HLS streaming.
func WithCDN(client CDNClient, baseURL string) Option {
	return func(s *Service) {
		s.cdn = client
		s.cdnBaseURL = baseURL
	}
}
