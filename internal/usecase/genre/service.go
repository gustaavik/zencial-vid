package genre

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles genre use cases.
type Service struct {
	genreRepo repository.GenreRepository
	log       *slog.Logger
}

// NewService creates a new genre Service.
func NewService(genreRepo repository.GenreRepository, log *slog.Logger) *Service {
	return &Service{genreRepo: genreRepo, log: log}
}
