package video

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
)

// Service handles video use cases.
type Service struct {
	videoRepo  repository.VideoRepository
	genreRepo  repository.GenreRepository
	subRepo    repository.SubscriptionRepository
	planRepo   repository.PlanRepository
	storage    storage.StorageService
	dispatcher event.Dispatcher
	log        *slog.Logger
}

// NewService creates a new video Service.
func NewService(
	videoRepo repository.VideoRepository,
	genreRepo repository.GenreRepository,
	subRepo repository.SubscriptionRepository,
	planRepo repository.PlanRepository,
	storage storage.StorageService,
	dispatcher event.Dispatcher,
	log *slog.Logger,
) *Service {
	return &Service{
		videoRepo:  videoRepo,
		genreRepo:  genreRepo,
		subRepo:    subRepo,
		planRepo:   planRepo,
		storage:    storage,
		dispatcher: dispatcher,
		log:        log,
	}
}
