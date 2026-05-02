package v1

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	authuc "github.com/zenfulcode/zencial/internal/usecase/auth"
	genreuc "github.com/zenfulcode/zencial/internal/usecase/genre"
	planuc "github.com/zenfulcode/zencial/internal/usecase/plan"
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
	Plan                 *planuc.Service
	Subscription         *subscriptionuc.Service
	Watchlist            *watchlistuc.Service
	WatchProgress        *watchprogressuc.Service
	TokenService         auth.TokenService
	Storage              storage.StorageService
	InternalSharedSecret string
	Log                  *slog.Logger
}

// RegisterRoutes registers all V1 API routes.
func RegisterRoutes(r chi.Router, deps *Deps) {
	authHandler := NewAuthHandler(deps.Auth)
	genreHandler := NewGenreHandler(deps.Genre)
	userHandler := NewUserHandler(deps.User)
	videoHandler := NewVideoHandler(deps.Video, deps.Subscription, deps.Storage)
	planHandler := NewPlanHandler(deps.Plan)
	subscriptionHandler := NewSubscriptionHandler(deps.Subscription)
	watchlistHandler := NewWatchlistHandler(deps.Watchlist, deps.Storage)
	watchProgressHandler := NewWatchProgressHandler(deps.WatchProgress, deps.Storage)
	transcodeCallbackHandler := NewTranscodeCallbackHandler(deps.Video)

	// Internal service-to-service routes (CDN callbacks). Outside the JWT chain.
	r.Route("/internal", func(r chi.Router) {
		r.Use(middleware.InternalAuth(deps.InternalSharedSecret))
		r.Post("/videos/{id}/transcode-callback", transcodeCallbackHandler.Handle)
	})

	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
	})

	// Public genre routes (read-only)
	r.Route("/genres", func(r chi.Router) {
		r.Get("/", genreHandler.List)
		r.Get("/{id}", genreHandler.GetByID)
	})

	// Public plan routes (active plans only)
	r.Get("/plans", planHandler.ListActive)

	// Public video routes with optional auth (for is_accessible field)
	r.Group(func(r chi.Router) {
		r.Use(middleware.OptionalAuthenticate(deps.TokenService))

		r.Get("/videos", videoHandler.ListPublished)
		r.Get("/videos/{id}", videoHandler.GetByID)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(deps.TokenService))

		// Auth (requires token)
		r.Post("/auth/logout", authHandler.Logout)

		// User profile (self)
		r.Get("/me", userHandler.GetMe)
		r.Put("/me", userHandler.UpdateMe)
		r.Delete("/me", userHandler.DeleteMe)

		// User subscription (self)
		r.Get("/me/subscription", subscriptionHandler.GetMySubscription)

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
			r.Post("/videos", videoHandler.Upload)
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
		})
	})
}
