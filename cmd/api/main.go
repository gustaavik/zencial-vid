package main

import (
	"context"
	logpac "log"
	"log/slog"
	"net/http"
	"os"
	"time"

	v1 "github.com/zenfulcode/zencial/internal/adapter/handler/v1"
	"github.com/zenfulcode/zencial/internal/adapter/messaging"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/infrastructure/buildinfo"
	"github.com/zenfulcode/zencial/internal/infrastructure/cdn"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
	"github.com/zenfulcode/zencial/internal/infrastructure/database"
	"github.com/zenfulcode/zencial/internal/infrastructure/logger"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/infrastructure/server"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/clock"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	audituc "github.com/zenfulcode/zencial/internal/usecase/audit"
	authuc "github.com/zenfulcode/zencial/internal/usecase/auth"
	billinguc "github.com/zenfulcode/zencial/internal/usecase/billing"
	genreuc "github.com/zenfulcode/zencial/internal/usecase/genre"
	planuc "github.com/zenfulcode/zencial/internal/usecase/plan"
	sessionuc "github.com/zenfulcode/zencial/internal/usecase/session"
	subscriptionuc "github.com/zenfulcode/zencial/internal/usecase/subscription"
	useruc "github.com/zenfulcode/zencial/internal/usecase/user"
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"
	watchlistuc "github.com/zenfulcode/zencial/internal/usecase/watchlist"
	watchprogressuc "github.com/zenfulcode/zencial/internal/usecase/watchprogress"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logpac.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	appLog := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(appLog)

	appLog.Info("starting Zencial VOD API",
		"version", buildinfo.Version,
		"commit", buildinfo.Commit,
		"build_time", buildinfo.BuildTime,
	)

	// Connect to PostgreSQL
	dbPool, err := database.NewPostgres(ctx, &cfg.Database)
	if err != nil {
		appLog.Error("failed to connect to database", "error", err)
		cancel()
		os.Exit(1) //nolint:gocritic
	}
	defer dbPool.Close()

	// Infrastructure services
	hasher := auth.NewBcryptHasher()
	tokens := auth.NewSessionTokenService()
	clk := clock.RealClock{}

	// Repositories
	userRepo := postgres.NewUserRepository(dbPool)
	sessionRepo := postgres.NewSessionRepository(dbPool)
	genreRepo := postgres.NewGenreRepository(dbPool)
	videoRepo := postgres.NewVideoRepository(dbPool)
	planRepo := postgres.NewPlanRepository(dbPool)
	subRepo := postgres.NewSubscriptionRepository(dbPool)
	watchlistRepo := postgres.NewWatchlistRepository(dbPool, videoRepo)
	watchProgressRepo := postgres.NewWatchProgressRepository(dbPool, videoRepo)
	auditLogRepo := postgres.NewAuditLogRepository(dbPool)

	// Event dispatcher
	dispatcher := messaging.NewEventDispatcher(appLog)

	// Storage (S3)
	storageService, err := storage.NewS3Service(&cfg.Storage)
	if err != nil {
		appLog.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	if err := storageService.EnsureBucket(ctx); err != nil {
		appLog.Error("failed to ensure storage bucket", "error", err)
		os.Exit(1)
	}
	appLog.Info("S3 storage initialized",
		"endpoint", cfg.Storage.Endpoint,
		"public_endpoint", cfg.Storage.PublicEndpoint,
		"bucket", cfg.Storage.Bucket,
		"region", cfg.Storage.Region,
	)

	// Authentication middleware
	authenticator := middleware.NewSessionAuthenticator(sessionRepo, userRepo, tokens, clk, cfg.Session, appLog)

	// Use cases
	authService := authuc.NewService(userRepo, sessionRepo, tokens, hasher, dispatcher, appLog, clk, cfg.Session)
	sessionService := sessionuc.NewService(sessionRepo, dispatcher, appLog, clk)
	genreService := genreuc.NewService(genreRepo, dispatcher, appLog)
	userService := useruc.NewService(userRepo, hasher, dispatcher, appLog)
	planService := planuc.NewService(planRepo, dispatcher, appLog)
	subscriptionService := subscriptionuc.NewService(subRepo, planRepo, dispatcher, appLog)
	billingService := billinguc.NewService(userRepo, planRepo, subRepo, billinguc.Config{
		SecretKey:     cfg.Stripe.SecretKey,
		WebhookSecret: cfg.Stripe.WebhookSecret,
		Currency:      cfg.Stripe.Currency,
	}, appLog)
	// Video service with optional CDN integration
	var videoOpts []videouc.Option
	if cfg.CDN.BaseURL != "" {
		// Use internal URL for backend→CDN calls, fall back to base URL.
		cdnInternalURL := cfg.CDN.InternalURL
		if cdnInternalURL == "" {
			cdnInternalURL = cfg.CDN.BaseURL
		}
		cdnClient := cdn.New(cdnInternalURL)
		videoOpts = append(videoOpts, videouc.WithCDN(cdnClient, cfg.CDN.BaseURL))
		appLog.Info("CDN integration enabled", "public_url", cfg.CDN.BaseURL, "internal_url", cdnInternalURL)
		if cfg.InternalAPI.SharedSecret == "" {
			appLog.Warn("CDN integration enabled but INTERNAL_API_SHARED_SECRET is unset — transcode-completion callbacks will be rejected")
		}
	}
	videoService := videouc.NewService(videoRepo, genreRepo, subRepo, planRepo, storageService, dispatcher, appLog, videoOpts...)
	watchlistService := watchlistuc.NewService(watchlistRepo, videoRepo, appLog)
	watchProgressService := watchprogressuc.NewService(watchProgressRepo, videoRepo, appLog)
	auditService := audituc.NewService(auditLogRepo, appLog)

	// Persist every dispatched domain event into the audit log.
	audituc.Register(dispatcher, auditLogRepo, appLog)

	// Background: purge expired/revoked sessions on a ticker. Cancelled when
	// ctx is cancelled (during graceful shutdown), so we don't race with
	// dbPool.Close().
	go runSessionCleanup(ctx, sessionService, cfg.Session.CleanupInterval, appLog)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery(appLog))
	r.Use(middleware.Logger(appLog))
	r.Use(middleware.CORS(cfg.Server))

	// API info
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		httputil.JSON(w, http.StatusOK, struct {
			Name      string `json:"name"`
			Version   string `json:"version"`
			Commit    string `json:"commit"`
			BuildTime string `json:"build_time"`
		}{
			Name:      "Zencial VOD API",
			Version:   buildinfo.Version,
			Commit:    buildinfo.Commit,
			BuildTime: buildinfo.BuildTime,
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		v1.RegisterRoutes(r, &v1.Deps{
			Auth:                 authService,
			Genre:                genreService,
			User:                 userService,
			Video:                videoService,
			Plan:                 planService,
			Subscription:         subscriptionService,
			Billing:              billingService,
			Watchlist:            watchlistService,
			WatchProgress:        watchProgressService,
			Audit:                auditService,
			Session:              sessionService,
			Authenticator:        authenticator,
			Storage:              storageService,
			InternalSharedSecret: cfg.InternalAPI.SharedSecret,
			Log:                  appLog,
		})
	})

	// Start server
	srv := server.New(cfg.Server, r, appLog)
	if err := srv.Start(); err != nil {
		appLog.Error("server error", "error", err)
		os.Exit(1)
	}
}

// runSessionCleanup periodically deletes expired/revoked session rows.
// Returns when ctx is cancelled.
func runSessionCleanup(ctx context.Context, svc *sessionuc.Service, interval time.Duration, log *slog.Logger) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := svc.PurgeExpired(ctx)
			if err != nil {
				log.Error("session cleanup failed", "error", err)
				continue
			}
			if n > 0 {
				log.Info("session cleanup", "deleted", n)
			}
		}
	}
}
