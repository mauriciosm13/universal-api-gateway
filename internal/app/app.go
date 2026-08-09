package app

import (
	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	middlewarestub "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/stub"
	observabilitystub "github.com/mauriciomendonca/universal-api-gateway/internal/observability/stub"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/server"
)

// New builds the gateway dependency graph from configuration and M0 stubs.
func New(cfg config.Config) server.Dependencies {
	return server.Dependencies{
		Config:        cfg,
		Authenticator: authstub.NewNoOpAuthenticator(),
		Router:        routingstub.NewNoOpRouter(),
		Limiter:       ratelimitstub.NewNoOpLimiter(),
		Pipeline:      middlewarestub.NewPassthroughPipeline(),
		Logger:        observabilitystub.NewNoOpLogger(),
		Tracer:        observabilitystub.NewNoOpTracer(),
	}
}
