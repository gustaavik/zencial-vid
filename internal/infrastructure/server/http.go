package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// Server wraps http.Server with graceful shutdown.
type Server struct {
	httpServer *http.Server
	cfg        config.ServerConfig
	log        *slog.Logger
}

// New creates a new HTTP server.
func New(cfg config.ServerConfig, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
		cfg: cfg,
		log: log,
	}
}

// Start begins listening and handles graceful shutdown on SIGINT/SIGTERM.
func (s *Server) Start() error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("starting server", "addr", s.cfg.Addr())
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		s.log.Info("shutting down server", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	s.log.Info("server stopped")
	return nil
}
