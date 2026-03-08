package v1

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	authuc "github.com/zenfulcode/zencial/internal/usecase/auth"
	cataloguc "github.com/zenfulcode/zencial/internal/usecase/catalog"
	contentuc "github.com/zenfulcode/zencial/internal/usecase/content"
	streaminguc "github.com/zenfulcode/zencial/internal/usecase/streaming"
	subscriptionuc "github.com/zenfulcode/zencial/internal/usecase/subscription"
	useruc "github.com/zenfulcode/zencial/internal/usecase/user"
	watchlistuc "github.com/zenfulcode/zencial/internal/usecase/watchlist"
)

// Deps holds all dependencies needed by V1 handlers.
type Deps struct {
	Auth         *authuc.Service
	User         *useruc.Service
	Content      *contentuc.Service
	Catalog      *cataloguc.Service
	Streaming    *streaminguc.Service
	Subscription *subscriptionuc.Service
	Watchlist    *watchlistuc.Service
	TokenService auth.TokenService
	Storage      storage.StorageService
	Log          *slog.Logger
}

// RegisterRoutes registers all V1 API routes.
func RegisterRoutes(r chi.Router, deps Deps) {
	authHandler := NewAuthHandler(deps.Auth)
	userHandler := NewUserHandler(deps.User)
	contentHandler := NewContentHandler(deps.Content)
	catalogHandler := NewCatalogHandler(deps.Catalog)
	streamingHandler := NewStreamingHandler(deps.Streaming)
	subscriptionHandler := NewSubscriptionHandler(deps.Subscription)
	watchlistHandler := NewWatchlistHandler(deps.Watchlist)

	var uploadHandler *UploadHandler
	if deps.Storage != nil {
		uploadHandler = NewUploadHandler(deps.Storage)
	}

	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
	})

	// Public: subscription plans
	r.Get("/plans", subscriptionHandler.ListPlans)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(deps.TokenService))

		// Auth (requires token)
		r.Post("/auth/logout", authHandler.Logout)

		// User profile
		r.Route("/users/me", func(r chi.Router) {
			r.Get("/", userHandler.GetMe)
			r.Patch("/", userHandler.UpdateMe)
			r.Delete("/", userHandler.DeleteMe)
		})

		// Content
		r.Route("/content", func(r chi.Router) {
			r.Get("/", contentHandler.List)
			r.Get("/featured", contentHandler.Featured)
			r.Get("/{slug}", contentHandler.GetBySlug)
			r.Get("/{slug}/seasons", contentHandler.GetSeasons)
			r.Get("/{slug}/seasons/{seasonNumber}/episodes", contentHandler.GetEpisodes)
		})

		// Catalog
		r.Get("/genres", catalogHandler.ListGenres)
		r.Get("/genres/{slug}/content", catalogHandler.ContentByGenre)
		r.Get("/categories", catalogHandler.ListCategories)
		r.Get("/search", contentHandler.Search)

		// Streaming
		r.Route("/streaming", func(r chi.Router) {
			r.Post("/sessions", streamingHandler.StartSession)
			r.Delete("/sessions/{id}", streamingHandler.EndSession)
			r.Put("/progress", streamingHandler.UpdateProgress)
			r.Get("/progress/{contentId}", streamingHandler.GetProgress)
			r.Get("/continue-watching", streamingHandler.ContinueWatching)
		})

		// Subscription
		r.Route("/subscriptions", func(r chi.Router) {
			r.Get("/me", subscriptionHandler.GetCurrent)
			r.Post("/", subscriptionHandler.Subscribe)
			r.Patch("/me/plan", subscriptionHandler.ChangePlan)
			r.Post("/me/cancel", subscriptionHandler.Cancel)
		})

		// Watchlist
		r.Route("/watchlist", func(r chi.Router) {
			r.Get("/", watchlistHandler.List)
			r.Post("/{contentId}", watchlistHandler.Add)
			r.Delete("/{contentId}", watchlistHandler.Remove)
			r.Get("/{contentId}/status", watchlistHandler.Status)
		})

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireRole(entity.RoleAdmin))

			if uploadHandler != nil {
				r.Post("/upload", uploadHandler.Upload)
				r.Post("/upload/init", uploadHandler.InitUpload)
			}

			r.Route("/content", func(r chi.Router) {
				r.Get("/", contentHandler.AdminList)
				r.Get("/{id}", contentHandler.AdminGetByID)
				r.Put("/{id}", contentHandler.Update)
				r.Delete("/{id}", contentHandler.Delete)
				r.Post("/{id}/publish", contentHandler.Publish)
				r.Post("/{id}/archive", contentHandler.Archive)
				r.Post("/{id}/asset", contentHandler.AttachVideoAsset)
			})

			r.Post("/films", contentHandler.CreateFilm)
			r.Post("/videos", contentHandler.CreateVideo)
			r.Post("/series", contentHandler.CreateSeries)

			r.Get("/users", userHandler.AdminListUsers)
			r.Patch("/users/{id}/status", userHandler.AdminUpdateStatus)

			r.Route("/subscriptions", func(r chi.Router) {
				r.Get("/", subscriptionHandler.AdminListSubscriptions)
				r.Post("/", subscriptionHandler.AdminCreateSubscription)
				r.Get("/user/{userId}", subscriptionHandler.AdminGetUserSubscription)
				r.Patch("/{id}/plan", subscriptionHandler.AdminChangePlan)
				r.Post("/{id}/reactivate", subscriptionHandler.AdminReactivateSubscription)
				r.Post("/{id}/cancel", subscriptionHandler.AdminCancelSubscription)
			})

			r.Route("/genres", func(r chi.Router) {
				r.Post("/", catalogHandler.CreateGenre)
				r.Put("/{id}", catalogHandler.UpdateGenre)
				r.Delete("/{id}", catalogHandler.DeleteGenre)
			})
		})
	})
}
