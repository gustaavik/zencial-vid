package content

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles content use cases.
type Service struct {
	contentRepo repository.ContentRepository
	catalogRepo repository.CatalogRepository
	log         *slog.Logger
}

// NewService creates a new content Service.
func NewService(contentRepo repository.ContentRepository, catalogRepo repository.CatalogRepository, log *slog.Logger) *Service {
	return &Service{contentRepo: contentRepo, catalogRepo: catalogRepo, log: log}
}
