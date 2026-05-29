package musiccue

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
)

// Service handles music cue use cases.
type Service struct {
	cueRepo   repository.MusicCueRepository
	videoRepo repository.VideoRepository
	storage   storage.StorageService
	log       *slog.Logger
}

// NewService creates a new music cue Service.
func NewService(
	cueRepo repository.MusicCueRepository,
	videoRepo repository.VideoRepository,
	storage storage.StorageService,
	log *slog.Logger,
) *Service {
	return &Service{
		cueRepo:   cueRepo,
		videoRepo: videoRepo,
		storage:   storage,
		log:       log,
	}
}
