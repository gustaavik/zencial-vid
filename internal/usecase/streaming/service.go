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

	// GetByID only loads the base Content row; we need Film/Video/Asset
	// sub-entities populated for IsPlayable() and resolveManifestURL().
	s.loadContentExtras(ctx, content)

	if !content.IsPlayable() {
		return nil, apperror.BadRequest(apperror.CodeContentNotPlayable, "content is not available for playback", domain.ErrContentNotPlayable)
	}

	session := entity.NewStreamSession(input.UserID, input.ContentID, input.EpisodeID, input.DeviceInfo, input.IPAddress)
	if err := s.streamingRepo.CreateSession(ctx, session); err != nil {
		s.log.Error("creating session", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create session", err)
	}

	manifestURL := s.resolveManifestURL(content, input.EpisodeID)

	return &StartSessionOutput{
		Session:     session,
		ManifestURL: manifestURL,
	}, nil
}

// resolveManifestURL builds the HLS manifest URL from the content asset.
// It prefers explicit rendition URLs, then falls back to constructing one
// from the CDN base URL and the asset storage key.
func (s *Service) resolveManifestURL(content *entity.Content, episodeID *uuid.UUID) string {
	// For episodes, find the matching episode asset
	if episodeID != nil && content.Series != nil {
		for _, season := range content.Series.Seasons {
			for _, ep := range season.Episodes {
				if ep.ID == *episodeID {
					return s.assetToURL(ep.Asset)
				}
			}
		}
	}

	switch content.Type {
	case entity.ContentTypeFilm:
		if content.Film != nil {
			return s.assetToURL(content.Film.Asset)
		}
	case entity.ContentTypeVideo:
		if content.Video != nil {
			return s.assetToURL(content.Video.Asset)
		}
	}

	// No asset URL could be resolved
	return ""
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

// loadContentExtras populates the Film/Video/Asset sub-entities on a Content
// that was loaded with the base repository GetByID (which only returns the row
// from the content table). Mirrors content.Service.loadContentExtras.
func (s *Service) loadContentExtras(ctx context.Context, content *entity.Content) {
	if content.Type == entity.ContentTypeVideo {
		video, err := s.contentRepo.GetVideoForContent(ctx, content.ID)
		if err != nil {
			s.log.Error("loading video data for streaming", "error", err)
		} else {
			content.Video = video
		}
	}

	if content.Type == entity.ContentTypeFilm || content.Type == entity.ContentTypeVideo {
		asset, err := s.contentRepo.GetVideoAssetForContent(ctx, content.ID)
		if err != nil {
			s.log.Error("loading video asset for streaming", "error", err)
		} else if asset != nil {
			if content.Type == entity.ContentTypeFilm {
				if content.Film == nil {
					content.Film = &entity.Film{ContentID: content.ID}
				}
				content.Film.Asset = *asset
			} else if content.Video != nil {
				content.Video.Asset = *asset
			}
		}
	}
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
