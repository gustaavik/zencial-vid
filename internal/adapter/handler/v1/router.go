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

		// Admin routes
		// r.Route("/admin", func(r chi.Router) {
		// 	r.Use(middleware.RequireRole(entity.RoleAdmin))

		// 	if uploadHandler != nil {
		// 		r.Post("/upload", uploadHandler.Upload)
		// 		r.Post("/upload/init", uploadHandler.InitUpload)
		// 	}

		// 	r.Route("/content", func(r chi.Router) {
		// 		r.Get("/", contentHandler.AdminList)
		// 		r.Get("/{id}", contentHandler.AdminGetByID)
		// 		r.Put("/{id}", contentHandler.Update)
		// 		r.Delete("/{id}", contentHandler.Delete)
		// 		r.Post("/{id}/publish", contentHandler.Publish)
		// 		r.Post("/{id}/archive", contentHandler.Archive)
		// 		r.Post("/{id}/asset", contentHandler.AttachVideoAsset)
		// 	})

		// 	r.Post("/films", contentHandler.CreateFilm)
		// 	r.Post("/videos", contentHandler.CreateVideo)
		// 	r.Post("/series", contentHandler.CreateSeries)

		// 	r.Get("/users", userHandler.AdminListUsers)
		// 	r.Patch("/users/{id}/status", userHandler.AdminUpdateStatus)

		// 	r.Route("/subscriptions", func(r chi.Router) {
		// 		r.Get("/", subscriptionHandler.AdminListSubscriptions)
		// 		r.Post("/", subscriptionHandler.AdminCreateSubscription)
		// 		r.Get("/user/{userId}", subscriptionHandler.AdminGetUserSubscription)
		// 		r.Patch("/{id}/plan", subscriptionHandler.AdminChangePlan)
		// 		r.Post("/{id}/reactivate", subscriptionHandler.AdminReactivateSubscription)
		// 		r.Post("/{id}/cancel", subscriptionHandler.AdminCancelSubscription)
		// 	})

		// 	r.Route("/genres", func(r chi.Router) {
		// 		r.Post("/", catalogHandler.CreateGenre)
		// 		r.Put("/{id}", catalogHandler.UpdateGenre)
		// 		r.Delete("/{id}", catalogHandler.DeleteGenre)
		// 	})

		// 	r.Route("/plans", func(r chi.Router) {
		// 		r.Get("/", subscriptionHandler.AdminListAllPlans)
		// 		r.Post("/", subscriptionHandler.AdminCreatePlan)
		// 		r.Put("/{id}", subscriptionHandler.AdminUpdatePlan)
		// 		r.Delete("/{id}", subscriptionHandler.AdminDeactivatePlan)
		// 	})
		// })
	})
}
