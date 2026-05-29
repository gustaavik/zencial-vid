package caption

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
)

// Service handles caption use cases.
type Service struct {
	captionRepo repository.CaptionRepository
	videoRepo   repository.VideoRepository
	storage     storage.StorageService
	log         *slog.Logger
}

// NewService creates a new caption Service.
func NewService(
	captionRepo repository.CaptionRepository,
	videoRepo repository.VideoRepository,
	storageSvc storage.StorageService,
	log *slog.Logger,
) *Service {
	return &Service{
		captionRepo: captionRepo,
		videoRepo:   videoRepo,
		storage:     storageSvc,
		log:         log,
	}
}
