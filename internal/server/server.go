package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mauriciomendonca/universal-api-gateway/internal/deps"
)

// Dependencies is an alias for the shared gateway dependency struct.
type Dependencies = deps.Gateway

// Server wraps the HTTP server and its routing.
type Server struct {
	httpServer *http.Server
	deps       Dependencies
}

// New creates a configured HTTP server with health endpoints.
func New(deps Dependencies) *Server {
	return &Server{
		deps: deps,
		httpServer: &http.Server{
			Addr:         deps.Config.Addr(),
			Handler:      newRoutes(deps),
			ReadTimeout:  deps.Config.ReadTimeout,
			WriteTimeout: deps.Config.WriteTimeout,
			IdleTimeout:  deps.Config.IdleTimeout,
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
