package server

import (
	"context"
	"fmt"
	"net/http"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	middlewareport "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
	observabilityport "github.com/mauriciomendonca/universal-api-gateway/internal/observability/port"
	ratelimitport "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/port"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

// Dependencies holds injected gateway ports and configuration.
type Dependencies struct {
	Config        config.Config
	Authenticator authport.Authenticator
	Router        routingport.Router
	Limiter       ratelimitport.Limiter
	Pipeline      middlewareport.Pipeline
	Logger        observabilityport.Logger
	Tracer        observabilityport.Tracer
}

// Server wraps the HTTP server and its routing.
type Server struct {
	httpServer *http.Server
	deps       Dependencies
}

// New creates a configured HTTP server with health endpoints.
func New(deps Dependencies) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /health/live", handleLive)
	mux.HandleFunc("GET /health/ready", handleReady)

	return &Server{
		deps: deps,
		httpServer: &http.Server{
			Addr:         deps.Config.Addr(),
			Handler:      mux,
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
