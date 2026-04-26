package mapper

import (
	"context"
	"log/slog"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"
)

// VideoToResponse maps a Video entity to a VideoResponse DTO.
func VideoToResponse(ctx context.Context, video *entity.Video, store storage.StorageService) dto.VideoResponse {
	genreIDs := make([]string, len(video.GenreIDs))
	for i, gid := range video.GenreIDs {
		genreIDs[i] = gid.String()
	}

	resp := dto.VideoResponse{
		ID:               video.ID.String(),
		Title:            video.Title,
		Slug:             video.Slug.String(),
		Description:      video.Description,
		Creator:          video.Creator,
		Duration:         video.Duration.Seconds,
		ContentRating:    video.ContentRating,
		Status:           string(video.Status),
		FileSize:         video.FileSize,
		GenreIDs:         genreIDs,
		MinimumPlanLevel: video.MinimumPlanLevel,
		TranscodeError:   video.TranscodeError,
		CreatedAt:        video.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        video.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if video.ThumbnailKey != "" && store != nil {
		url, err := store.PresignedGetURL(ctx, video.ThumbnailKey, storage.DefaultThumbnailURLExpiry)
		if err != nil {
			slog.Error("generating thumbnail presigned URL", "key", video.ThumbnailKey, "error", err)
		} else {
			resp.ThumbnailURL = url
		}
	}
	return resp
}

// VideoToResponseWithAccess maps a Video entity to a VideoResponse DTO with access info.
func VideoToResponseWithAccess(ctx context.Context, video *entity.Video, store storage.StorageService, userPlanLevel *int) dto.VideoResponse {
	resp := VideoToResponse(ctx, video, store)
	var accessible bool
	switch {
	case !video.RequiresSubscription():
		accessible = true
	case userPlanLevel != nil:
		accessible = *userPlanLevel >= *video.MinimumPlanLevel
	default:
		accessible = false
	}
	resp.IsAccessible = &accessible
	return resp
}

// VideosToResponse maps a slice of Video entities to VideoResponse DTOs.
func VideosToResponse(ctx context.Context, videos []entity.Video, store storage.StorageService) []dto.VideoResponse {
	result := make([]dto.VideoResponse, len(videos))
	for i := range videos {
		result[i] = VideoToResponse(ctx, &videos[i], store)
	}
	return result
}

// VideosToResponseWithAccess maps a slice of Video entities to VideoResponse DTOs with access info.
func VideosToResponseWithAccess(ctx context.Context, videos []entity.Video, store storage.StorageService, userPlanLevel *int) []dto.VideoResponse {
	result := make([]dto.VideoResponse, len(videos))
	for i := range videos {
		result[i] = VideoToResponseWithAccess(ctx, &videos[i], store, userPlanLevel)
	}
	return result
}

// BulkResultToResponse maps a BulkResult to a BulkResultResponse DTO.
func BulkResultToResponse(result *videouc.BulkResult) dto.BulkResultResponse {
	succeeded := make([]string, len(result.Succeeded))
	for i, id := range result.Succeeded {
		succeeded[i] = id.String()
	}

	failed := make([]dto.BulkFailureResponse, len(result.Failed))
	for i, f := range result.Failed {
		failed[i] = dto.BulkFailureResponse{
			ID:    f.ID.String(),
			Error: f.Error,
		}
	}

	return dto.BulkResultResponse{
		Succeeded: succeeded,
		Failed:    failed,
	}
}

// StreamToResponse maps a StreamOutput to a VideoStreamResponse DTO.
func StreamToResponse(output *videouc.StreamOutput) dto.VideoStreamResponse {
	return dto.VideoStreamResponse{
		URL:       output.URL,
		ExpiresAt: output.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		Type:      output.Type,
	}
}
