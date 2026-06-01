package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// PreflightItem represents a single checklist item for submission readiness.
type PreflightItem struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Passed  bool   `json:"passed"`
	Blocker bool   `json:"blocker"`
}

// PreflightResult is the structured pre-flight checklist response.
type PreflightResult struct {
	VideoID    string          `json:"video_id"`
	ReadyCount int             `json:"ready_count"`
	TotalCount int             `json:"total_count"`
	Blockers   int             `json:"blockers"`
	Items      []PreflightItem `json:"items"`
}

// Preflight returns the submission readiness checklist for a video.
func (s *Service) Preflight(ctx context.Context, videoID, uploaderID uuid.UUID) (*PreflightResult, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", nil)
	}
	if video.UploadedBy != uploaderID {
		return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil)
	}

	items := buildChecklist(ctx, video, s)

	ready := 0
	blockers := 0
	for _, item := range items {
		if item.Passed {
			ready++
		}
		if !item.Passed && item.Blocker {
			blockers++
		}
	}

	return &PreflightResult{
		VideoID:    videoID.String(),
		ReadyCount: ready,
		TotalCount: len(items),
		Blockers:   blockers,
		Items:      items,
	}, nil
}

func buildChecklist(ctx context.Context, v *entity.Video, s *Service) []PreflightItem {
	items := make([]PreflightItem, 0, 8)
	items = append(items,
		PreflightItem{
			Key:     "file_encoded",
			Label:   "File uploaded & encoded",
			Passed:  v.Status == entity.VideoStatusPublished || v.Status == entity.VideoStatusProcessing,
			Blocker: true,
		},
		PreflightItem{
			Key:     "thumbnail",
			Label:   "Thumbnail picked",
			Passed:  v.ThumbnailKey != "",
			Blocker: false,
		},
		PreflightItem{
			Key:     "title_description",
			Label:   "Title & description",
			Passed:  v.Title != "" && v.Description != "",
			Blocker: true,
		},
		PreflightItem{
			Key:     "genres_rating",
			Label:   "Genres & rating",
			Passed:  len(v.GenreIDs) > 0 && v.ContentRating != "",
			Blocker: false,
		},
		PreflightItem{
			Key:     "visibility",
			Label:   "Visibility chosen",
			Passed:  v.Visibility != "",
			Blocker: true,
		},
		PreflightItem{
			Key:     "monetization",
			Label:   "Monetization selected",
			Passed:  len(v.MonetizationTypes) > 0,
			Blocker: false,
		},
	)

	// Music rights — check for blocking cues.
	musicPassed := true
	if s.musicCueRepo != nil {
		hasBlocking, err := s.musicCueRepo.HasBlockingCues(ctx, v.ID)
		if err == nil {
			musicPassed = !hasBlocking
		}
	}
	items = append(items, PreflightItem{
		Key:     "music_rights",
		Label:   "Music rights confirmed",
		Passed:  musicPassed,
		Blocker: true,
		// Series membership.
	}, PreflightItem{
		Key:    "series",
		Label:  "Belongs to series",
		Passed: v.SeriesID != nil,
	})

	return items
}
