package v1

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	analyticsuc "github.com/zenfulcode/zencial/internal/usecase/analytics"
	audituc "github.com/zenfulcode/zencial/internal/usecase/audit"
	authuc "github.com/zenfulcode/zencial/internal/usecase/auth"
	billinguc "github.com/zenfulcode/zencial/internal/usecase/billing"
	captionuc "github.com/zenfulcode/zencial/internal/usecase/caption"
	castuc "github.com/zenfulcode/zencial/internal/usecase/cast"
	chapteruc "github.com/zenfulcode/zencial/internal/usecase/chapter"
	genreuc "github.com/zenfulcode/zencial/internal/usecase/genre"
	musiccueuc "github.com/zenfulcode/zencial/internal/usecase/musiccue"
	planuc "github.com/zenfulcode/zencial/internal/usecase/plan"
	seasonuc "github.com/zenfulcode/zencial/internal/usecase/season"
	seriesuc "github.com/zenfulcode/zencial/internal/usecase/series"
	sessionuc "github.com/zenfulcode/zencial/internal/usecase/session"
	subscriptionuc "github.com/zenfulcode/zencial/internal/usecase/subscription"
	useruc "github.com/zenfulcode/zencial/internal/usecase/user"
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"
	watchlistuc "github.com/zenfulcode/zencial/internal/usecase/watchlist"
	watchprogressuc "github.com/zenfulcode/zencial/internal/usecase/watchprogress"
)

// Deps holds all dependencies needed by V1 handlers.
type Deps struct {
	Auth                 *authuc.Service
	Genre                *genreuc.Service
	User                 *useruc.Service
	Video                *videouc.Service
	Series               *seriesuc.Service
	Season               *seasonuc.Service
	Chapter              *chapteruc.Service
	Caption              *captionuc.Service
	MusicCue             *musiccueuc.Service
	Plan                 *planuc.Service
	Subscription         *subscriptionuc.Service
	Billing              *billinguc.Service
	Watchlist            *watchlistuc.Service
	WatchProgress        *watchprogressuc.Service
	Audit                *audituc.Service
	Session              *sessionuc.Service
	Analytics            *analyticsuc.Service
	Cast                 *castuc.Service
	Authenticator        *middleware.SessionAuthenticator
	Storage              storage.StorageService
	CDNURLs              mapper.ThumbnailURLBuilder
	InternalSharedSecret string
	Log                  *slog.Logger
}

// RegisterRoutes registers all V1 API routes.
func RegisterRoutes(r chi.Router, deps *Deps) {
	authHandler := NewAuthHandler(deps.Auth)
	genreHandler := NewGenreHandler(deps.Genre)
	userHandler := NewUserHandler(deps.User)
	videoHandler := NewVideoHandler(deps.Video, deps.Subscription, deps.Storage, deps.CDNURLs)
	seriesHandler := NewSeriesHandler(deps.Series, deps.CDNURLs)
	seasonHandler := NewSeasonHandler(deps.Season)
	chapterHandler := NewChapterHandler(deps.Chapter)
	captionHandler := NewCaptionHandler(deps.Caption)
	musicCueHandler := NewMusicCueHandler(deps.MusicCue)
	planHandler := NewPlanHandler(deps.Plan)
	subscriptionHandler := NewSubscriptionHandler(deps.Subscription)
	billingHandler := NewBillingHandler(deps.Billing)
	watchlistHandler := NewWatchlistHandler(deps.Watchlist, deps.CDNURLs)
	watchProgressHandler := NewWatchProgressHandler(deps.WatchProgress, deps.CDNURLs)
	transcodeCallbackHandler := NewTranscodeCallbackHandler(deps.Video)
	auditLogHandler := NewAuditLogHandler(deps.Audit)
	sessionHandler := NewSessionHandler(deps.Session)
	castHandler := NewCastHandler(deps.Cast, deps.CDNURLs)
	analyticsHandler := NewAnalyticsHandler(deps.Analytics)

	// Internal service-to-service routes (CDN callbacks). Outside the session chain.
	r.Route("/internal", func(r chi.Router) {
		r.Use(middleware.InternalAuth(deps.InternalSharedSecret))
		r.Post("/videos/{id}/transcode-callback", transcodeCallbackHandler.Handle)
	})

	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Public genre routes (read-only)
	r.Route("/genres", func(r chi.Router) {
		r.Get("/", genreHandler.List)
		r.Get("/{id}", genreHandler.GetByID)
	})

	// Public plan routes (active plans only)
	r.Get("/plans", planHandler.ListActive)

	// Stripe webhooks. Stripe signs requests, so this route stays outside session auth.
	r.Post("/billing/webhook", billingHandler.HandleWebhook)

	// Public video routes with optional auth (for is_accessible field)
	r.Group(func(r chi.Router) {
		r.Use(deps.Authenticator.OptionalAuthenticate)

		r.Get("/videos", videoHandler.ListPublished)
		r.Get("/videos/featured", videoHandler.GetFeatured)
		r.Get("/videos/{id}", videoHandler.GetByID)

		// Cast is public (anyone can see who's in a video, or all videos for a cast member)
		r.Get("/videos/{id}/cast", castHandler.List)
		r.Get("/cast/{id}/videos", castHandler.ListVideos)

		// Series (public read)
		r.Get("/series", seriesHandler.ListPublished)
		r.Get("/series/{id}", seriesHandler.GetByID)
		r.Get("/series/{id}/episodes", seriesHandler.ListEpisodes)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(deps.Authenticator.Authenticate)

		// Auth (requires session)
		r.Post("/auth/logout", authHandler.Logout)

		// Session management (self)
		r.Get("/me/sessions", sessionHandler.ListMine)
		r.Delete("/me/sessions/{sessionID}", sessionHandler.RevokeMine)
		r.Post("/me/sessions/revoke-others", sessionHandler.RevokeOthers)

		// User profile (self)
		r.Get("/me", userHandler.GetMe)
		r.Put("/me", userHandler.UpdateMe)
		r.Delete("/me", userHandler.DeleteMe)
		r.Get("/me/handle/check", userHandler.CheckHandle)

		// User subscription (self)
		r.Get("/me/subscription", subscriptionHandler.GetMySubscription)
		r.Post("/billing/checkout-sessions", billingHandler.CreateCheckoutSession)
		r.Post("/billing/portal-sessions", billingHandler.CreatePortalSession)
		r.Get("/billing/invoices", billingHandler.ListInvoices)

		// Watchlist (self)
		r.Get("/me/watchlist", watchlistHandler.List)
		r.Get("/me/watchlist/{video_id}", watchlistHandler.GetStatus)
		r.Post("/me/watchlist/{video_id}", watchlistHandler.Add)
		r.Delete("/me/watchlist/{video_id}", watchlistHandler.Remove)

		// Watch progress / continue watching (self)
		r.Get("/me/watch-progress", watchProgressHandler.List)
		r.Get("/me/watch-progress/{video_id}", watchProgressHandler.Get)
		r.Put("/me/watch-progress/{video_id}", watchProgressHandler.Upsert)
		r.Delete("/me/watch-progress/{video_id}", watchProgressHandler.Delete)

		// Video streaming (any authenticated user)
		r.Get("/videos/{id}/stream", videoHandler.Stream)

		// Playback analytics ingestion (cumulative heartbeats from players)
		r.With(middleware.RateLimit(2, 10)).
			Post("/videos/{id}/playback-events", analyticsHandler.RecordPlayback)

		// Series watch progress (any authenticated user)
		r.Get("/series/{id}/next-episode", seriesHandler.GetNextEpisode)
		r.Put("/series/{id}/watch-progress", seriesHandler.UpdateWatchProgress)
		r.Get("/series/{id}/watch-progress", seriesHandler.GetWatchProgress)

		// Publisher + Admin routes
		r.Route("/publisher", func(r chi.Router) {
			r.Use(middleware.RequireAnyRole(entity.RolePublisher, entity.RoleAdmin))

			// Video management (own videos only for publishers)
			r.Get("/videos", videoHandler.ListOwned)
			r.Post("/videos/uploads", videoHandler.InitiateUpload)
			r.Post("/videos", videoHandler.CompleteUpload)
			r.Put("/videos/{id}", videoHandler.Update)
			r.Put("/videos/{id}/thumbnail", videoHandler.UploadThumbnail)
			r.Post("/videos/{id}/publish", videoHandler.PublishOwned)
			r.Delete("/videos/{id}", videoHandler.DeleteOwned)
			r.Get("/videos/{id}/preflight", videoHandler.Preflight)
			r.Post("/videos/{id}/submit", videoHandler.Submit)

			// Chapters
			r.Get("/videos/{id}/chapters", chapterHandler.List)
			r.Put("/videos/{id}/chapters", chapterHandler.Replace)
			r.Delete("/videos/{id}/chapters/{chapterID}", chapterHandler.Delete)

			// Captions
			r.Get("/videos/{id}/captions", captionHandler.List)
			r.Post("/videos/{id}/captions", captionHandler.InitiateUpload)
			r.Post("/videos/{id}/captions/register", captionHandler.Register)
			r.Delete("/videos/{id}/captions/{lang}", captionHandler.Delete)
			r.Post("/videos/{id}/captions/{lang}/publish", captionHandler.Publish)

			// Music cues
			r.Get("/videos/{id}/music-cues", musicCueHandler.List)
			r.Post("/videos/{id}/music-cues", musicCueHandler.Create)
			r.Put("/videos/{id}/music-cues/{cueID}", musicCueHandler.Update)
			r.Delete("/videos/{id}/music-cues/{cueID}", musicCueHandler.Delete)
			r.Post("/videos/{id}/music-cues/{cueID}/clearance", musicCueHandler.InitiateClearanceUpload)
			r.Post("/videos/{id}/music-cues/{cueID}/clearance/complete", musicCueHandler.CompleteClearanceUpload)

			// Analytics
			r.Get("/videos/{id}/analytics", analyticsHandler.VideoStats)
			r.Get("/analytics/summary", analyticsHandler.Summary)

			// Series management (own series)
			r.Post("/series", seriesHandler.Create)
			r.Put("/series/{id}", seriesHandler.Update)
			r.Post("/series/{id}/episodes", seriesHandler.AddEpisode)
			r.Delete("/series/{id}/episodes/{videoID}", seriesHandler.RemoveEpisode)
			r.Get("/publisher/series", seriesHandler.ListOwned)
			r.Post("/publisher/series/{id}/publish", seriesHandler.PublishOwned)
			r.Delete("/publisher/series/{id}", seriesHandler.ArchiveOwned)

			// Seasons
			r.Get("/series/{id}/seasons", seasonHandler.List)
			r.Post("/series/{id}/seasons", seasonHandler.Create)
			r.Put("/series/{id}/seasons/{seasonID}", seasonHandler.Update)
			r.Delete("/series/{id}/seasons/{seasonID}", seasonHandler.Delete)
		})

		// Cast management (publisher or admin)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAnyRole(entity.RolePublisher, entity.RoleAdmin))
			r.Get("/admin/cast", castHandler.ListAll)
			r.Post("/videos/{id}/cast", castHandler.Create)
			r.Put("/videos/{videoID}/cast/{creditID}", castHandler.UpdateCredit)
			r.Delete("/videos/{videoID}/cast/{creditID}", castHandler.DeleteFromVideo)
			r.Put("/cast/{id}", castHandler.UpdateCast)
			r.Put("/cast/{id}/picture", castHandler.UploadPicture)
			r.Delete("/cast/{id}", castHandler.Delete)
			r.Post("/cast/{id}/unarchive", castHandler.Unarchive)
		})

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(entity.RoleAdmin))

			// Genre management
			r.Post("/genres", genreHandler.Create)
			r.Put("/genres/{id}", genreHandler.Update)
			r.Delete("/genres/{id}", genreHandler.Delete)

			// Bulk genre operations
			r.Post("/admin/genres/bulk-create", genreHandler.BulkCreate)
			r.Post("/admin/genres/bulk-delete", genreHandler.BulkDelete)

			// Plan management
			r.Post("/plans", planHandler.Create)
			r.Get("/plans/{id}", planHandler.GetByID)
			r.Put("/plans/{id}", planHandler.Update)
			r.Delete("/plans/{id}", planHandler.Delete)
			r.Get("/admin/plans", planHandler.List)

			// Subscription management
			r.Post("/admin/subscriptions", subscriptionHandler.Assign)
			r.Delete("/admin/subscriptions/{id}", subscriptionHandler.Cancel)

			// Video management
			r.Post("/videos/uploads", videoHandler.InitiateUpload)
			r.Post("/videos", videoHandler.CompleteUpload)
			r.Put("/videos/{id}", videoHandler.Update)
			r.Put("/videos/{id}/thumbnail", videoHandler.UploadThumbnail)
			r.Post("/videos/{id}/publish", videoHandler.Publish)
			r.Post("/videos/{id}/unarchive", videoHandler.Unarchive)
			r.Delete("/videos/{id}", videoHandler.Delete)

			// Admin video listing (all statuses)
			r.Get("/admin/videos", videoHandler.ListAll)

			// Bulk video operations
			r.Post("/admin/videos/bulk-publish", videoHandler.BulkPublish)
			r.Post("/admin/videos/bulk-archive", videoHandler.BulkDelete)
			r.Post("/admin/videos/bulk-unarchive", videoHandler.BulkUnarchive)

			// Maintenance
			r.Post("/admin/videos/purge-orphans", videoHandler.PurgeOrphans)

			// Featured
			r.Post("/admin/videos/{id}/feature", videoHandler.SetFeatured)
			r.Delete("/admin/videos/{id}/feature", videoHandler.UnsetFeatured)

			// Moderation
			r.Get("/admin/moderation/queue", videoHandler.ModerationQueue)
			r.Post("/admin/videos/{id}/approve", videoHandler.ApproveSubmission)
			r.Post("/admin/videos/{id}/reject", videoHandler.RejectSubmission)

			// Admin analytics (any video)
			r.Get("/admin/videos/{id}/analytics", analyticsHandler.VideoStats)
			r.Get("/admin/analytics/summary", analyticsHandler.AdminSummary)

			// Series management (admin)
			r.Get("/admin/series", seriesHandler.AdminListAll)
			r.Post("/series/{id}/publish", seriesHandler.AdminPublish)
			r.Delete("/series/{id}", seriesHandler.AdminArchive)
			r.Post("/series/{id}/unarchive", seriesHandler.AdminUnarchive)

			// Audit log (admin)
			r.Get("/admin/audit-logs", auditLogHandler.List)

			// User management (admin)
			r.Get("/admin/users", userHandler.ListUsers)
			r.Post("/admin/users", userHandler.CreateUser)
			r.Get("/admin/users/{id}", userHandler.GetUser)
			r.Put("/admin/users/{id}", userHandler.UpdateUser)
			r.Delete("/admin/users/{id}", userHandler.DeleteUser)
			r.Put("/admin/users/{id}/status", userHandler.UpdateUserStatus)
			r.Get("/admin/users/{id}/subscriptions", subscriptionHandler.ListByUser)
			r.Get("/admin/users/{id}/watchlist", watchlistHandler.ListByUser)
			r.Get("/admin/users/{id}/watch-progress", watchProgressHandler.ListByUser)

			// Session management (admin)
			r.Get("/admin/users/{id}/sessions", sessionHandler.AdminListByUser)
			r.Post("/admin/users/{id}/sessions/revoke-all", sessionHandler.AdminRevokeAll)
			r.Delete("/admin/sessions/{sessionID}", sessionHandler.AdminRevoke)
		})
	})
}
