package content

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AttachVideoAsset creates and links a video asset to the given content.
func (s *Service) AttachVideoAsset(ctx context.Context, contentID uuid.UUID, storageKey string) (*entity.VideoAsset, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		s.log.Error("getting content for asset attach", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}

	// Only film and video types support video assets.
	if content.Type != entity.ContentTypeFilm && content.Type != entity.ContentTypeVideo {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "only film and video content types support video assets", nil)
	}

	// Default to "ready" — no transcoding pipeline exists yet.
	// When a processing pipeline is added, this should revert to VideoAssetPending
	// and transition to VideoAssetReady after transcoding completes.
	asset := &entity.VideoAsset{
		ID:         uuid.New(),
		StorageKey: storageKey,
		Status:     entity.VideoAssetReady,
	}

	if err := s.contentRepo.CreateVideoAsset(ctx, asset, contentID); err != nil {
		s.log.Error("creating video asset", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create video asset", err)
	}

	return asset, nil
}

// GetVideoAsset retrieves the video asset for the given content.
func (s *Service) GetVideoAsset(ctx context.Context, contentID uuid.UUID) (*entity.VideoAsset, *apperror.AppError) {
	asset, err := s.contentRepo.GetVideoAssetForContent(ctx, contentID)
	if err != nil {
		s.log.Error("getting video asset", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video asset", err)
	}
	return asset, nil
}
