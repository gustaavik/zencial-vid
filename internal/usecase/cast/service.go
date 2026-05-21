package cast

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
)

// Service handles cast use cases.
type Service struct {
	castRepo  repository.CastRepository
	videoRepo repository.VideoRepository
	storage   storage.StorageService
	log       *slog.Logger
}

// NewService creates a new cast Service.
func NewService(castRepo repository.CastRepository, videoRepo repository.VideoRepository, log *slog.Logger, s storage.StorageService) *Service {
	svc := &Service{castRepo: castRepo, videoRepo: videoRepo, log: log, storage: s}
	return svc
}

// resolvePictureURL populates c.PictureURL from c.PictureKey using the storage backend.
func (s *Service) resolvePictureURL(c *entity.Cast) {
	if s.storage != nil && c.PictureKey != "" {
		c.PictureURL = s.storage.PublicURL(c.PictureKey)
	}
}
