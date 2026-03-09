package v1

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	authuc "github.com/zenfulcode/zencial/internal/usecase/auth"
)

// Deps holds all dependencies needed by V1 handlers.
type Deps struct {
	Auth         *authuc.Service
	TokenService auth.TokenService
	Storage      storage.StorageService
	Log          *slog.Logger
}

// RegisterRoutes registers all V1 API routes.
func RegisterRoutes(r chi.Router, deps Deps) {
	authHandler := NewAuthHandler(deps.Auth)

	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(deps.TokenService))

		// Auth (requires token)
		r.Post("/auth/logout", authHandler.Logout)
	})
}
