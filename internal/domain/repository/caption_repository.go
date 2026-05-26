package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// CaptionRepository defines persistence operations for video captions.
type CaptionRepository interface {
	Upsert(ctx context.Context, caption *entity.Caption) error
	GetByVideoAndLang(ctx context.Context, videoID uuid.UUID, languageCode string) (*entity.Caption, error)
	ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.Caption, error)
	Update(ctx context.Context, caption *entity.Caption) error
	DeleteByVideoAndLang(ctx context.Context, videoID uuid.UUID, languageCode string) error
}
