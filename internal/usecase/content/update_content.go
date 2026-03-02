package content

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateContentInput holds the optional fields for updating content.
type UpdateContentInput struct {
	Title       *string
	Description *string
	Synopsis    *string
	Rating      *string
	ReleaseYear *int
	PosterURL   *string
	BackdropURL *string
	TrailerURL  *string
	Director    *string
	IsFeatured  *bool
	CreatorName *string
	IsFree      *bool
}

// Update applies partial updates to existing content.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateContentInput) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting content for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}

	if input.Title != nil && *input.Title != content.Title {
		content.Title = *input.Title
		baseSlug, err := valueobject.NewSlug(*input.Title)
		if err == nil {
			exists, existsErr := s.contentRepo.ExistsBySlug(ctx, baseSlug.String())
			if existsErr != nil {
				s.log.Error("checking slug existence", "error", existsErr)
			} else if exists {
				content.Slug = baseSlug.WithRandomID()
			} else {
				content.Slug = baseSlug
			}
		}
	}
	if input.Description != nil {
		content.Description = *input.Description
	}
	if input.Synopsis != nil {
		content.Synopsis = *input.Synopsis
	}
	if input.Rating != nil {
		content.Rating = valueobject.ContentRating(*input.Rating)
	}
	if input.ReleaseYear != nil {
		content.ReleaseYear = *input.ReleaseYear
	}
	if input.PosterURL != nil {
		content.PosterURL = *input.PosterURL
	}
	if input.BackdropURL != nil {
		content.BackdropURL = *input.BackdropURL
	}
	if input.TrailerURL != nil {
		content.TrailerURL = *input.TrailerURL
	}
	if input.Director != nil {
		content.Director = *input.Director
	}
	if input.IsFeatured != nil {
		content.IsFeatured = *input.IsFeatured
	}
	content.UpdatedAt = time.Now()

	if err := s.contentRepo.Update(ctx, content); err != nil {
		s.log.Error("updating content", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update content", err)
	}

	// Update video-specific fields if applicable.
	if content.Type == entity.ContentTypeVideo && (input.CreatorName != nil || input.IsFree != nil) {
		video, err := s.contentRepo.GetVideoForContent(ctx, id)
		if err != nil {
			s.log.Error("getting video for update", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video record", err)
		}
		if video != nil {
			if input.CreatorName != nil {
				video.CreatorName = *input.CreatorName
			}
			if input.IsFree != nil {
				video.IsFree = *input.IsFree
			}
			if err := s.contentRepo.UpdateVideo(ctx, video); err != nil {
				s.log.Error("updating video record", "error", err)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to update video record", err)
			}
			content.Video = video
		}
	}

	return content, nil
}
