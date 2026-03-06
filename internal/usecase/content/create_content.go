package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CreateContentInput holds the data required to create new content.
type CreateContentInput struct {
	Type        string
	Title       string
	Description string
	Synopsis    string
	Rating      string
	ReleaseYear int
	PosterURL   string
	BackdropURL string
	TrailerURL  string
	Director    string
	CreatorName string
	IsFree      *bool
}

// Create creates new content from the given input.
func (s *Service) Create(ctx context.Context, input CreateContentInput) (*entity.Content, *apperror.AppError) {
	baseSlug, err := valueobject.NewSlug(input.Title)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid title for slug", err)
	}

	// Ensure slug uniqueness by appending a random ID if the base slug is taken.
	slug := baseSlug
	exists, err := s.contentRepo.ExistsBySlug(ctx, slug.String())
	if err != nil {
		s.log.Error("checking slug existence", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check slug", err)
	}
	if exists {
		slug = baseSlug.WithRandomID()
	}

	// Default rating to G for videos if not provided.
	rating := valueobject.ContentRating(input.Rating)
	if input.Rating == "" && input.Type == string(entity.ContentTypeVideo) {
		rating = valueobject.RatingG
	}

	now := time.Now()
	content := &entity.Content{
		ID:          uuid.New(),
		Type:        entity.ContentType(input.Type),
		Title:       input.Title,
		Slug:        slug,
		Description: input.Description,
		Synopsis:    input.Synopsis,
		Rating:      rating,
		ReleaseYear: input.ReleaseYear,
		PosterURL:   input.PosterURL,
		BackdropURL: input.BackdropURL,
		TrailerURL:  input.TrailerURL,
		Director:    input.Director,
		Status:      entity.ContentStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.contentRepo.Create(ctx, content); err != nil {
		if errors.Is(err, domain.ErrSlugAlreadyExists) {
			return nil, apperror.Conflict(apperror.CodeSlugConflict, "content with this slug already exists", err)
		}
		s.log.Error("creating content", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create content", err)
	}

	// Create video-specific record.
	if content.Type == entity.ContentTypeVideo {
		isFree := false
		if input.IsFree != nil {
			isFree = *input.IsFree
		}
		video := &entity.Video{
			ContentID:   content.ID,
			CreatorName: input.CreatorName,
			IsFree:      isFree,
		}
		if err := s.contentRepo.CreateVideo(ctx, video); err != nil {
			s.log.Error("creating video record", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to create video record", err)
		}
		content.Video = video
	}

	return content, nil
}
