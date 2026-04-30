package watchlist

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Service handles watchlist use cases.
type Service struct {
	watchlistRepo repository.WatchlistRepository
	videoRepo     repository.VideoRepository
	log           *slog.Logger
}

// NewService creates a new watchlist Service.
func NewService(watchlistRepo repository.WatchlistRepository, videoRepo repository.VideoRepository, log *slog.Logger) *Service {
	return &Service{
		watchlistRepo: watchlistRepo,
		videoRepo:     videoRepo,
		log:           log,
	}
}

// Add inserts (userID, videoID) into the user's watchlist.
//
// Idempotent: re-adding an existing entry is a no-op and returns nil.
// Returns NotFound when the video does not exist.
func (s *Service) Add(ctx context.Context, userID, videoID uuid.UUID) *apperror.AppError {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("watchlist: getting video", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if err := s.watchlistRepo.Add(ctx, userID, videoID); err != nil {
		s.log.Error("watchlist: adding entry", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to add to watchlist", err)
	}
	return nil
}

// Remove deletes (userID, videoID) from the user's watchlist.
//
// Returns NotFound when no such entry exists.
func (s *Service) Remove(ctx context.Context, userID, videoID uuid.UUID) *apperror.AppError {
	if err := s.watchlistRepo.Remove(ctx, userID, videoID); err != nil {
		if errors.Is(err, domain.ErrWatchlistEntryNotFound) {
			return apperror.NotFound(apperror.CodeWatchlistEntryNotFound, "watchlist entry not found", err)
		}
		s.log.Error("watchlist: removing entry", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to remove from watchlist", err)
	}
	return nil
}

// IsInWatchlist returns whether the given video is in the user's watchlist.
func (s *Service) IsInWatchlist(ctx context.Context, userID, videoID uuid.UUID) (bool, *apperror.AppError) {
	exists, err := s.watchlistRepo.Exists(ctx, userID, videoID)
	if err != nil {
		s.log.Error("watchlist: checking existence", "error", err)
		return false, apperror.Internal(apperror.CodeInternalError, "failed to check watchlist", err)
	}
	return exists, nil
}

// List returns the user's watchlisted videos sorted by added_at DESC.
func (s *Service) List(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.watchlistRepo.ListVideos(ctx, userID, page)
	if err != nil {
		s.log.Error("watchlist: listing videos", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list watchlist", err)
	}
	return videos, total, nil
}
