package series

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles series business logic.
type Service struct {
	seriesRepo   repository.SeriesRepository
	seriesWpRepo repository.SeriesWatchProgressRepository
	videoRepo    repository.VideoRepository
	genreRepo    repository.GenreRepository
	dispatcher   event.Dispatcher
	log          *slog.Logger
}

// NewService creates a new series Service.
func NewService(
	seriesRepo repository.SeriesRepository,
	seriesWpRepo repository.SeriesWatchProgressRepository,
	videoRepo repository.VideoRepository,
	genreRepo repository.GenreRepository,
	dispatcher event.Dispatcher,
	log *slog.Logger,
) *Service {
	return &Service{
		seriesRepo:   seriesRepo,
		seriesWpRepo: seriesWpRepo,
		videoRepo:    videoRepo,
		genreRepo:    genreRepo,
		dispatcher:   dispatcher,
		log:          log,
	}
}
