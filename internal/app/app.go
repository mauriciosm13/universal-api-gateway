package app

import (
	"github.com/mauriciomendonca/universal-api-gateway/internal/auth"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	"github.com/mauriciomendonca/universal-api-gateway/internal/middleware"
	"github.com/mauriciomendonca/universal-api-gateway/internal/observability"
	"github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing"
	"github.com/mauriciomendonca/universal-api-gateway/internal/server"
)

// DefaultModules returns the standard M0 module set for gateway wiring.
func DefaultModules() []di.Module {
	return []di.Module{
		auth.Module{},
		routing.Module{},
		ratelimit.Module{},
		middleware.Module{},
		observability.Module{},
		server.Module{},
	}
}

// Build constructs server.Dependencies from configuration and registered modules.
// When no modules are passed, DefaultModules is used.
func Build(cfg config.Config, modules ...di.Module) (server.Dependencies, error) {
	if len(modules) == 0 {
		modules = DefaultModules()
	}
	return di.NewBuilder(cfg, modules...).Build()
}
