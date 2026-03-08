package content

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// ContentDetail is a discriminated union returned by typed use-case methods.
// Exactly one of Film, Video, or Series is non-nil.
type ContentDetail struct {
	Type   entity.ContentType
	Film   *entity.Film
	Video  *entity.Video
	Series *entity.Series
}

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
