package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	v1 "github.com/zenfulcode/zencial/internal/adapter/handler/v1"
	"github.com/zenfulcode/zencial/internal/adapter/messaging"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
	"github.com/zenfulcode/zencial/internal/infrastructure/database"
	"github.com/zenfulcode/zencial/internal/infrastructure/logger"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/redis"
	"github.com/zenfulcode/zencial/internal/infrastructure/server"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	authuc "github.com/zenfulcode/zencial/internal/usecase/auth"
	genreuc "github.com/zenfulcode/zencial/internal/usecase/genre"
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/zenfulcode/zencial/docs"
)

// @title           Zencial VOD API
// @version         1.0
// @description     Video on Demand streaming platform API
// @termsOfService  http://zencial.com/terms/

// @contact.name   Zencial API Support
// @contact.email  support@zencial.com

// @license.name  Proprietary
// @license.url   http://zencial.com/license/

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format: Bearer {token}
func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	// Connect to PostgreSQL
	dbPool, err := database.NewPostgres(ctx, cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Connect to Redis
	redisClient, err := database.NewRedis(ctx, cfg.Redis)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Infrastructure services
	tokenService := auth.NewJWTService(cfg.JWT)
	hasher := auth.NewBcryptHasher()

	// Repositories
	userRepo := postgres.NewUserRepository(dbPool)
	genreRepo := postgres.NewGenreRepository(dbPool)
	videoRepo := postgres.NewVideoRepository(dbPool)

	// Redis stores
	sessionStore := redis.NewSessionStore(redisClient, cfg.JWT.RefreshDuration)

	// Event dispatcher
	dispatcher := messaging.NewEventDispatcher(log)

	// Storage (MinIO)
	storageService, err := storage.NewMinIOService(cfg.Storage)
	if err != nil {
		log.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	if err := storageService.EnsureBucket(ctx); err != nil {
		log.Error("failed to ensure storage bucket", "error", err)
		os.Exit(1)
	}

	// Use cases
	authService := authuc.NewService(userRepo, tokenService, hasher, sessionStore, dispatcher, log)
	genreService := genreuc.NewService(genreRepo, log)
	videoService := videouc.NewService(videoRepo, genreRepo, storageService, dispatcher, log)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery(log))
	r.Use(middleware.Logger(log))
	r.Use(middleware.CORS(cfg.Server))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		v1.RegisterRoutes(r, v1.Deps{
			Auth:         authService,
			Genre:        genreService,
			Video:        videoService,
			TokenService: tokenService,
			Storage:      storageService,
			Log:          log,
		})
	})

	// Start server
	srv := server.New(cfg.Server, r, log)
	if err := srv.Start(); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
