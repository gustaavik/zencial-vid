package season

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles season use cases.
type Service struct {
	seasonRepo repository.SeasonRepository
	seriesRepo repository.SeriesRepository
	log        *slog.Logger
}

// NewService creates a new season Service.
func NewService(
	seasonRepo repository.SeasonRepository,
	seriesRepo repository.SeriesRepository,
	log *slog.Logger,
) *Service {
	return &Service{
		seasonRepo: seasonRepo,
		seriesRepo: seriesRepo,
		log:        log,
	}
}
