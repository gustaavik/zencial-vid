package cast

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles cast use cases.
type Service struct {
	castRepo  repository.CastRepository
	videoRepo repository.VideoRepository
	log       *slog.Logger
}

// NewService creates a new cast Service.
func NewService(castRepo repository.CastRepository, videoRepo repository.VideoRepository, log *slog.Logger) *Service {
	return &Service{castRepo: castRepo, videoRepo: videoRepo, log: log}
}
