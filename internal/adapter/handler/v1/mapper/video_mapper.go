package mapper

import (
	"context"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"
)

// ThumbnailURLBuilder produces the public CDN URL for a video thumbnail.
// Defined here, in the mapper package, to keep the mapper free of
// infrastructure imports — *cdn.Client satisfies it implicitly.
type ThumbnailURLBuilder interface {
	ThumbnailURL(videoID string) string
}

// VideoToResponse maps a Video entity to a VideoResponse DTO.
func VideoToResponse(_ context.Context, video *entity.Video, urls ThumbnailURLBuilder) dto.VideoResponse {
	genreIDs := make([]string, len(video.GenreIDs))
	for i, gid := range video.GenreIDs {
		genreIDs[i] = gid.String()
	}

	resp := dto.VideoResponse{
		ID:                    video.ID.String(),
		Title:                 video.Title,
		Slug:                  video.Slug.String(),
		Description:           video.Description,
		Logline:               video.Logline,
		Creator:               video.Creator,
		Duration:              video.Duration.Seconds,
		ContentRating:         video.ContentRating,
		PrimaryLanguage:       video.PrimaryLanguage,
		Status:                string(video.Status),
		Visibility:            string(video.Visibility),
		FileSize:              video.FileSize,
		GenreIDs:              genreIDs,
		MinimumPlanLevel:      video.MinimumPlanLevel,
		TranscodeError:        video.TranscodeError,
		SeasonNumber:          video.SeasonNumber,
		EpisodeNumber:         video.EpisodeNumber,
		MonetizationTypes:     video.MonetizationTypes,
		PPVPriceCents:         video.PPVPriceCents,
		FreePreviewSeconds:    video.FreePreviewSeconds,
		AdBreakPositions:      video.AdBreakPositions,
		GeoRestrictionType:    string(video.GeoRestrictionType),
		GeoRestrictionRegions: video.GeoRestrictionRegions,
		RequireSignin:         video.RequireSignin,
		SubmissionStatus:      string(video.SubmissionStatus),
		ModeratorNotes:        video.ModeratorNotes,
		CreatedAt:             video.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:             video.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if video.SeriesID != nil {
		s := video.SeriesID.String()
		resp.SeriesID = &s
	}
	if video.ThumbnailKey != "" && urls != nil {
		resp.ThumbnailURL = urls.ThumbnailURL(video.ID.String())
	}
	if video.ScheduledPublishAt != nil {
		t := video.ScheduledPublishAt.UTC().Format("2006-01-02T15:04:05Z")
		resp.ScheduledPublishAt = &t
	}
	if video.SubmittedAt != nil {
		t := video.SubmittedAt.UTC().Format("2006-01-02T15:04:05Z")
		resp.SubmittedAt = &t
	}
	return resp
}

// VideoToResponseWithAccess maps a Video entity to a VideoResponse DTO with access info.
func VideoToResponseWithAccess(ctx context.Context, video *entity.Video, urls ThumbnailURLBuilder, userPlanLevel *int) dto.VideoResponse {
	resp := VideoToResponse(ctx, video, urls)
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
func VideosToResponse(ctx context.Context, videos []entity.Video, urls ThumbnailURLBuilder) []dto.VideoResponse {
	result := make([]dto.VideoResponse, len(videos))
	for i := range videos {
		result[i] = VideoToResponse(ctx, &videos[i], urls)
	}
	return result
}

// VideosToResponseWithAccess maps a slice of Video entities to VideoResponse DTOs with access info.
func VideosToResponseWithAccess(ctx context.Context, videos []entity.Video, urls ThumbnailURLBuilder, userPlanLevel *int) []dto.VideoResponse {
	result := make([]dto.VideoResponse, len(videos))
	for i := range videos {
		result[i] = VideoToResponseWithAccess(ctx, &videos[i], urls, userPlanLevel)
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

// PurgeOrphansToResponse maps a PurgeOrphansOutput to a PurgeOrphansResponse DTO.
func PurgeOrphansToResponse(out *videouc.PurgeOrphansOutput, dryRun bool) dto.PurgeOrphansResponse {
	dbOrphans := make([]string, len(out.DBOrphans))
	for i, id := range out.DBOrphans {
		dbOrphans[i] = id.String()
	}

	s3Orphans := out.S3Orphans
	if s3Orphans == nil {
		s3Orphans = []string{}
	}

	return dto.PurgeOrphansResponse{
		DryRun:    dryRun,
		DBOrphans: dbOrphans,
		S3Orphans: s3Orphans,
	}
}
