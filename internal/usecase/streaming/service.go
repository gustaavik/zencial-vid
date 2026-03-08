package streaming

import (
	"context"
	"errors"
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
	cdnBaseURL       string
	log              *slog.Logger
}

func NewService(
	streamingRepo repository.StreamingRepository,
	contentRepo repository.ContentRepository,
	subscriptionRepo repository.SubscriptionRepository,
	cdnBaseURL string,
	log *slog.Logger,
) *Service {
	return &Service{
		streamingRepo:    streamingRepo,
		contentRepo:      contentRepo,
		subscriptionRepo: subscriptionRepo,
		cdnBaseURL:       cdnBaseURL,
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

	// Clean up any existing sessions for the same user + content.
	// Handles browser refresh / tab close where the frontend cleanup doesn't fire.
	if err := s.streamingRepo.EndSessionsForContent(ctx, input.UserID, input.ContentID); err != nil {
		s.log.Warn("failed to clean up previous sessions", "error", err)
	}

	activeSessions, err := s.streamingRepo.GetActiveSessionsByUser(ctx, input.UserID)
	if err != nil {
		s.log.Error("getting active sessions", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check sessions", err)
	}
	if sub.Plan != nil && len(activeSessions) >= sub.Plan.MaxStreams {
		return nil, apperror.Conflict(apperror.CodeMaxStreamsReached, "maximum concurrent streams reached", domain.ErrMaxStreamsReached)
	}

	base, manifestURL, appErr := s.loadForStreaming(ctx, input.ContentID, input.EpisodeID)
	if appErr != nil {
		return nil, appErr
	}

	if !base.IsPlayable() {
		return nil, apperror.BadRequest(apperror.CodeContentNotPlayable, "content is not available for playback", domain.ErrContentNotPlayable)
	}

	session := entity.NewStreamSession(input.UserID, input.ContentID, input.EpisodeID, input.DeviceInfo, input.IPAddress)
	if err := s.streamingRepo.CreateSession(ctx, session); err != nil {
		s.log.Error("creating session", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create session", err)
	}

	return &StartSessionOutput{
		Session:     session,
		ManifestURL: manifestURL,
	}, nil
}

// loadForStreaming fetches the BaseContent and resolves the manifest URL for a content item.
func (s *Service) loadForStreaming(ctx context.Context, contentID uuid.UUID, episodeID *uuid.UUID) (*entity.BaseContent, string, *apperror.AppError) {
	ct, err := s.contentRepo.GetTypeByID(ctx, contentID)
	if err != nil {
		if errors.Is(err, domain.ErrContentNotFound) {
			return nil, "", apperror.NotFound(apperror.CodeContentNotFound, "content not found", err)
		}
		s.log.Error("getting content type for streaming", "error", err)
		return nil, "", apperror.Internal(apperror.CodeInternalError, "failed to get content", err)
	}

	switch ct {
	case entity.ContentTypeFilm:
		film, err := s.contentRepo.GetFilmByID(ctx, contentID)
		if err != nil || film == nil {
			return nil, "", apperror.NotFound(apperror.CodeContentNotFound, "film not found", domain.ErrContentNotFound)
		}
		url := ""
		if film.Asset != nil {
			url = s.assetToURL(*film.Asset)
		}
		return &film.BaseContent, url, nil

	case entity.ContentTypeVideo:
		video, err := s.contentRepo.GetVideoByID(ctx, contentID)
		if err != nil || video == nil {
			return nil, "", apperror.NotFound(apperror.CodeContentNotFound, "video not found", domain.ErrContentNotFound)
		}
		url := ""
		if video.Asset != nil {
			url = s.assetToURL(*video.Asset)
		}
		return &video.BaseContent, url, nil

	case entity.ContentTypeSeries:
		series, err := s.contentRepo.GetSeriesByID(ctx, contentID)
		if err != nil || series == nil {
			return nil, "", apperror.NotFound(apperror.CodeContentNotFound, "series not found", domain.ErrContentNotFound)
		}
		if !series.IsPublished() {
			return nil, "", apperror.BadRequest(apperror.CodeContentNotPlayable, "content is not available for playback", domain.ErrContentNotPlayable)
		}
		url := s.resolveEpisodeURL(ctx, contentID, episodeID)
		if url == "" {
			return nil, "", apperror.BadRequest(apperror.CodeContentNotPlayable, "episode not available for playback", domain.ErrContentNotPlayable)
		}
		// Synthesise a BaseContent with a ready asset so the IsPlayable() check in StartSession passes.
		base := &entity.BaseContent{
			ID:     series.ID,
			Type:   entity.ContentTypeSeries,
			Status: series.Status,
			Asset:  &entity.VideoAsset{Status: entity.VideoAssetReady},
		}
		return base, url, nil

	default:
		return nil, "", apperror.Internal(apperror.CodeInternalError, "unknown content type", nil)
	}
}

// resolveEpisodeURL finds the asset URL for a specific episode.
func (s *Service) resolveEpisodeURL(ctx context.Context, seriesID uuid.UUID, episodeID *uuid.UUID) string {
	if episodeID == nil {
		return ""
	}
	ep, err := s.contentRepo.GetEpisodeByID(ctx, *episodeID)
	if err != nil || ep == nil {
		s.log.Warn("episode not found for manifest URL", "episodeID", episodeID)
		return ""
	}
	return s.assetToURL(ep.Asset)
}

// assetToURL returns the best available URL for a video asset.
// It prefers explicit rendition URLs over constructing from StorageKey.
func (s *Service) assetToURL(asset entity.VideoAsset) string {
	// Prefer explicit rendition URLs (HD > FHD > others)
	for _, quality := range []string{"FHD", "HD", "SD", "UHD"} {
		for _, r := range asset.Qualities {
			if string(r.Quality) == quality && r.URL != "" {
				return r.URL
			}
		}
	}
	// First available rendition URL
	for _, r := range asset.Qualities {
		if r.URL != "" {
			return r.URL
		}
	}
	// Construct from storage key — serves the file directly (MP4 or HLS manifest)
	// via CDN_BASE_URL which points to the CDN or MinIO public endpoint + bucket.
	if asset.StorageKey != "" && s.cdnBaseURL != "" {
		return s.cdnBaseURL + "/" + asset.StorageKey
	}
	return ""
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
