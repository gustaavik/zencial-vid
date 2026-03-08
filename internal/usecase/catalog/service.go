package catalog

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

type Service struct {
	catalogRepo repository.CatalogRepository
	contentRepo repository.ContentRepository
	log         *slog.Logger
}

func NewService(catalogRepo repository.CatalogRepository, contentRepo repository.ContentRepository, log *slog.Logger) *Service {
	return &Service{catalogRepo: catalogRepo, contentRepo: contentRepo, log: log}
}

func (s *Service) ListGenres(ctx context.Context) ([]entity.Genre, *apperror.AppError) {
	genres, err := s.catalogRepo.ListGenres(ctx)
	if err != nil {
		s.log.Error("listing genres", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list genres", err)
	}
	return genres, nil
}

func (s *Service) ListCategories(ctx context.Context) ([]entity.Category, *apperror.AppError) {
	categories, err := s.catalogRepo.ListCategories(ctx)
	if err != nil {
		s.log.Error("listing categories", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list categories", err)
	}
	return categories, nil
}

func (s *Service) ContentByGenre(ctx context.Context, genreSlug string, page, perPage int) ([]entity.ContentSummary, int64, *apperror.AppError) {
	genre, err := s.catalogRepo.GetGenreBySlug(ctx, genreSlug)
	if err != nil {
		s.log.Error("getting genre by slug", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to get genre", err)
	}
	if genre == nil {
		return nil, 0, apperror.NotFound(apperror.CodeGenreNotFound, "genre not found", domain.ErrGenreNotFound)
	}

	contents, total, err := s.contentRepo.GetByGenre(ctx, genre.ID, page, perPage)
	if err != nil {
		s.log.Error("getting content by genre", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}
	return contents, total, nil
}

func (s *Service) CreateGenre(ctx context.Context, name string) (*entity.Genre, *apperror.AppError) {
	slug, err := valueobject.NewSlug(name)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid name for slug", err)
	}
	genre := &entity.Genre{
		ID:   uuid.New(),
		Name: name,
		Slug: slug.String(),
	}
	if err := s.catalogRepo.CreateGenre(ctx, genre); err != nil {
		s.log.Error("creating genre", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create genre", err)
	}
	return genre, nil
}

func (s *Service) UpdateGenre(ctx context.Context, id uuid.UUID, name string) (*entity.Genre, *apperror.AppError) {
	genre, err := s.catalogRepo.GetGenreByID(ctx, id)
	if err != nil || genre == nil {
		return nil, apperror.NotFound(apperror.CodeGenreNotFound, "genre not found", domain.ErrGenreNotFound)
	}
	slug, slugErr := valueobject.NewSlug(name)
	if slugErr != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid name for slug", slugErr)
	}
	genre.Name = name
	genre.Slug = slug.String()
	if err := s.catalogRepo.UpdateGenre(ctx, genre); err != nil {
		s.log.Error("updating genre", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update genre", err)
	}
	return genre, nil
}

func (s *Service) DeleteGenre(ctx context.Context, id uuid.UUID) *apperror.AppError {
	if err := s.catalogRepo.DeleteGenre(ctx, id); err != nil {
		s.log.Error("deleting genre", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete genre", err)
	}
	return nil
}
