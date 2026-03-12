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
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"
)

// Deps holds all dependencies needed by V1 handlers.
type Deps struct {
	Auth         *authuc.Service
	Genre        *genreuc.Service
	Video        *videouc.Service
	TokenService auth.TokenService
	Storage      storage.StorageService
	Log          *slog.Logger
}

// RegisterRoutes registers all V1 API routes.
func RegisterRoutes(r chi.Router, deps Deps) {
	authHandler := NewAuthHandler(deps.Auth)
	genreHandler := NewGenreHandler(deps.Genre)
	videoHandler := NewVideoHandler(deps.Video, deps.Storage)

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

	// Public video routes (published videos only)
	r.Route("/videos", func(r chi.Router) {
		r.Get("/", videoHandler.ListPublished)
		r.Get("/{id}", videoHandler.GetByID)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(deps.TokenService))

		// Auth (requires token)
		r.Post("/auth/logout", authHandler.Logout)

		// Video streaming (any authenticated user)
		r.Get("/videos/{id}/stream", videoHandler.Stream)

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(entity.RoleAdmin))

			// Genre management
			r.Post("/genres", genreHandler.Create)
			r.Put("/genres/{id}", genreHandler.Update)
			r.Delete("/genres/{id}", genreHandler.Delete)

			// Video management
			r.Post("/videos", videoHandler.Upload)
			r.Put("/videos/{id}", videoHandler.Update)
			r.Post("/videos/{id}/publish", videoHandler.Publish)
			r.Post("/videos/{id}/unarchive", videoHandler.Unarchive)
			r.Delete("/videos/{id}", videoHandler.Delete)

			// Admin video listing (all statuses)
			r.Get("/admin/videos", videoHandler.ListAll)
		})
	})
}
