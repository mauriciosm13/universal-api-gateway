package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
)

// Server wraps the HTTP server and its routing.
type Server struct {
	httpServer *http.Server
}

// New creates a configured HTTP server with health endpoints.
func New(cfg config.Config) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /health/live", handleLive)
	mux.HandleFunc("GET /health/ready", handleReady)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Start begins accepting HTTP connections.
func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("start server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
