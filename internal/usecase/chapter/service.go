package chapter

import (
	"log/slog"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Service handles chapter use cases.
type Service struct {
	chapterRepo repository.ChapterRepository
	videoRepo   repository.VideoRepository
	log         *slog.Logger
}

// NewService creates a new chapter Service.
func NewService(
	chapterRepo repository.ChapterRepository,
	videoRepo repository.VideoRepository,
	log *slog.Logger,
) *Service {
	return &Service{
		chapterRepo: chapterRepo,
		videoRepo:   videoRepo,
		log:         log,
	}
}
