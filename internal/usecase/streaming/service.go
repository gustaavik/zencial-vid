package streaming

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
	streamingRepo    repository.StreamingRepository
	contentRepo      repository.ContentRepository
	subscriptionRepo repository.SubscriptionRepository
	log              *slog.Logger
}

func NewService(
	streamingRepo repository.StreamingRepository,
	contentRepo repository.ContentRepository,
	subscriptionRepo repository.SubscriptionRepository,
	log *slog.Logger,
) *Service {
	return &Service{
		streamingRepo:    streamingRepo,
		contentRepo:      contentRepo,
		subscriptionRepo: subscriptionRepo,
		log:              log,
	}
}

type StartSessionInput struct {
	UserID     uuid.UUID
	ContentID  uuid.UUID
	EpisodeID  *uuid.UUID
	DeviceInfo string
	IPAddress  string
}

type StartSessionOutput struct {
	Session     *entity.StreamSession
	ManifestURL string
}

func (s *Service) StartSession(ctx context.Context, input StartSessionInput) (*StartSessionOutput, *apperror.AppError) {
	sub, err := s.subscriptionRepo.GetActiveByUserID(ctx, input.UserID)
	if err != nil {
		s.log.Error("getting subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check subscription", err)
	}
	if sub == nil || !sub.IsAccessible() {
		return nil, apperror.Forbidden(apperror.CodeNoActiveSubscription, "active subscription required", domain.ErrNoActiveSubscription)
	}

	activeSessions, err := s.streamingRepo.GetActiveSessionsByUser(ctx, input.UserID)
	if err != nil {
		s.log.Error("getting active sessions", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check sessions", err)
	}
	if sub.Plan != nil && len(activeSessions) >= sub.Plan.MaxStreams {
		return nil, apperror.Conflict(apperror.CodeMaxStreamsReached, "maximum concurrent streams reached", domain.ErrMaxStreamsReached)
	}

	content, err := s.contentRepo.GetByID(ctx, input.ContentID)
	if err != nil || content == nil {
		return nil, apperror.NotFound(apperror.CodeContentNotFound, "content not found", domain.ErrContentNotFound)
	}
	if !content.IsPlayable() {
		return nil, apperror.BadRequest(apperror.CodeContentNotPlayable, "content is not available for playback", domain.ErrContentNotPlayable)
	}

	session := entity.NewStreamSession(input.UserID, input.ContentID, input.EpisodeID, input.DeviceInfo, input.IPAddress)
	if err := s.streamingRepo.CreateSession(ctx, session); err != nil {
		s.log.Error("creating session", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create session", err)
	}

	// In production, this would generate a signed CDN URL
	manifestURL := "https://cdn.zencial.com/placeholder/master.m3u8"

	return &StartSessionOutput{
		Session:     session,
		ManifestURL: manifestURL,
	}, nil
}

func (s *Service) EndSession(ctx context.Context, userID, sessionID uuid.UUID) *apperror.AppError {
	if err := s.streamingRepo.EndSession(ctx, sessionID); err != nil {
		s.log.Error("ending session", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to end session", err)
	}
	return nil
}

type UpdateProgressInput struct {
	UserID    uuid.UUID
	ContentID uuid.UUID
	EpisodeID *uuid.UUID
	Position  int64
	Duration  int64
}

func (s *Service) UpdateProgress(ctx context.Context, input UpdateProgressInput) *apperror.AppError {
	progress := &entity.PlaybackProgress{
		ID:        uuid.New(),
		UserID:    input.UserID,
		ContentID: input.ContentID,
		EpisodeID: input.EpisodeID,
		Position:  input.Position,
		Duration:  input.Duration,
	}
	progress.MarkCompleted()

	if err := s.streamingRepo.SaveProgress(ctx, progress); err != nil {
		s.log.Error("saving progress", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to save progress", err)
	}
	return nil
}

func (s *Service) GetProgress(ctx context.Context, userID, contentID uuid.UUID, episodeID *uuid.UUID) (*entity.PlaybackProgress, *apperror.AppError) {
	progress, err := s.streamingRepo.GetProgress(ctx, userID, contentID, episodeID)
	if err != nil {
		s.log.Error("getting progress", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get progress", err)
	}
	return progress, nil
}

func (s *Service) ContinueWatching(ctx context.Context, userID uuid.UUID, limit int) ([]entity.ContinueWatching, *apperror.AppError) {
	items, err := s.streamingRepo.GetContinueWatching(ctx, userID, limit)
	if err != nil {
		s.log.Error("getting continue watching", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get continue watching", err)
	}
	return items, nil
}
