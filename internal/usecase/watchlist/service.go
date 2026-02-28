package watchlist

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

type Service struct {
	watchlistRepo repository.WatchlistRepository
	contentRepo   repository.ContentRepository
	log           *slog.Logger
}

func NewService(watchlistRepo repository.WatchlistRepository, contentRepo repository.ContentRepository, log *slog.Logger) *Service {
	return &Service{watchlistRepo: watchlistRepo, contentRepo: contentRepo, log: log}
}

func (s *Service) Add(ctx context.Context, userID, contentID uuid.UUID) *apperror.AppError {
	exists, err := s.watchlistRepo.Exists(ctx, userID, contentID)
	if err != nil {
		s.log.Error("checking watchlist", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to check watchlist", err)
	}
	if exists {
		return apperror.Conflict(apperror.CodeAlreadyInWatchlist, "already in watchlist", domain.ErrAlreadyInWatchlist)
	}

	item := entity.NewWatchlistItem(userID, contentID)
	if err := s.watchlistRepo.Add(ctx, item); err != nil {
		s.log.Error("adding to watchlist", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to add to watchlist", err)
	}
	return nil
}

func (s *Service) Remove(ctx context.Context, userID, contentID uuid.UUID) *apperror.AppError {
	if err := s.watchlistRepo.Remove(ctx, userID, contentID); err != nil {
		s.log.Error("removing from watchlist", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to remove from watchlist", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, page, perPage int) ([]entity.WatchlistItem, int64, *apperror.AppError) {
	items, total, err := s.watchlistRepo.GetByUserID(ctx, userID, page, perPage)
	if err != nil {
		s.log.Error("listing watchlist", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list watchlist", err)
	}
	return items, total, nil
}

func (s *Service) Status(ctx context.Context, userID, contentID uuid.UUID) (bool, *apperror.AppError) {
	exists, err := s.watchlistRepo.Exists(ctx, userID, contentID)
	if err != nil {
		s.log.Error("checking watchlist status", "error", err)
		return false, apperror.Internal(apperror.CodeInternalError, "failed to check watchlist", err)
	}
	return exists, nil
}
