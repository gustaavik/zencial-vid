package video

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
)

// CDNClient triggers HLS transcoding on the CDN service.
type CDNClient interface {
	TriggerTranscode(videoID string) error
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
