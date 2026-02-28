package content

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

type Service struct {
	contentRepo repository.ContentRepository
	catalogRepo repository.CatalogRepository
	log         *slog.Logger
}

func NewService(contentRepo repository.ContentRepository, catalogRepo repository.CatalogRepository, log *slog.Logger) *Service {
	return &Service{contentRepo: contentRepo, catalogRepo: catalogRepo, log: log}
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetBySlug(ctx, slug)
	if err != nil {
		s.log.Error("getting content by slug", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}
	return content, nil
}

func (s *Service) List(ctx context.Context, criteria entity.SearchCriteria) ([]entity.Content, int64, *apperror.AppError) {
	contents, total, err := s.contentRepo.Search(ctx, criteria)
	if err != nil {
		s.log.Error("listing content", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list content", err)
	}
	return contents, total, nil
}

func (s *Service) Search(ctx context.Context, criteria entity.SearchCriteria) ([]entity.Content, int64, *apperror.AppError) {
	return s.List(ctx, criteria)
}

func (s *Service) Featured(ctx context.Context, limit int) ([]entity.Content, *apperror.AppError) {
	contents, err := s.contentRepo.GetFeatured(ctx, limit)
	if err != nil {
		s.log.Error("getting featured content", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get featured content", err)
	}
	return contents, nil
}

func (s *Service) GetSeasons(ctx context.Context, slug string) ([]entity.Season, *apperror.AppError) {
	content, appErr := s.GetBySlug(ctx, slug)
	if appErr != nil {
		return nil, appErr
	}
	seasons, err := s.contentRepo.GetSeasonsForContent(ctx, content.ID)
	if err != nil {
		s.log.Error("getting seasons", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get seasons", err)
	}
	return seasons, nil
}

func (s *Service) GetEpisodes(ctx context.Context, slug string, seasonNumber int) ([]entity.Episode, *apperror.AppError) {
	content, appErr := s.GetBySlug(ctx, slug)
	if appErr != nil {
		return nil, appErr
	}
	seasons, err := s.contentRepo.GetSeasonsForContent(ctx, content.ID)
	if err != nil {
		s.log.Error("getting seasons for episodes", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get seasons", err)
	}
	for _, season := range seasons {
		if season.Number == seasonNumber {
			episodes, err := s.contentRepo.GetEpisodesForSeason(ctx, season.ID)
			if err != nil {
				s.log.Error("getting episodes", "error", err)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to get episodes", err)
			}
			return episodes, nil
		}
	}
	return nil, apperror.NotFound(apperror.CodeSeasonNotFound, "season not found", domain.ErrSeasonNotFound)
}

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
}

func (s *Service) Create(ctx context.Context, input CreateContentInput) (*entity.Content, *apperror.AppError) {
	slug, err := valueobject.NewSlug(input.Title)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid title for slug", err)
	}

	now := time.Now()
	content := &entity.Content{
		ID:          uuid.New(),
		Type:        entity.ContentType(input.Type),
		Title:       input.Title,
		Slug:        slug,
		Description: input.Description,
		Synopsis:    input.Synopsis,
		Rating:      valueobject.ContentRating(input.Rating),
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
		s.log.Error("creating content", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create content", err)
	}

	return content, nil
}

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
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateContentInput) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting content for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	if content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}

	if input.Title != nil {
		content.Title = *input.Title
		slug, err := valueobject.NewSlug(*input.Title)
		if err == nil {
			content.Slug = slug
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
	return content, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	if err := s.contentRepo.Delete(ctx, id); err != nil {
		s.log.Error("deleting content", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete content", err)
	}
	return nil
}

func (s *Service) Publish(ctx context.Context, id uuid.UUID) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil || content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}
	content.Publish()
	if err := s.contentRepo.Update(ctx, content); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to publish content", err)
	}
	return content, nil
}

func (s *Service) Archive(ctx context.Context, id uuid.UUID) (*entity.Content, *apperror.AppError) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil || content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}
	content.Archive()
	if err := s.contentRepo.Update(ctx, content); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to archive content", err)
	}
	return content, nil
}
