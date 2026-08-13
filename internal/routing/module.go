package routing

import (
	"fmt"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	pathrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/path"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

// Module registers routing port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	cfg := b.Config()

	if len(cfg.PathRoutes) > 0 {
		router, err := pathrouter.NewRouter(cfg.PathRoutes, cfg.DefaultUpstream)
		if err != nil {
			panic(fmt.Sprintf("routing: path router: %v", err))
		}

		b.ProvideRouter(router)
		return
	}

	if cfg.DefaultUpstream == "" {
		b.ProvideRouter(routingstub.NewNoOpRouter())
		return
	}

	router, err := staticrouter.NewRouter(cfg.DefaultUpstream)
	if err != nil {
		panic(fmt.Sprintf("routing: %v", err))
	}

	b.ProvideRouter(router)
}

func buildRouter(cfg config.Config) (routingport.Router, error) {
	routers := make([]routingport.Router, 0, 3)

	if len(cfg.HeaderRoutes) > 0 {
		headerRouter, err := headerrouter.NewRouter(cfg.HeaderRoutes)
		if err != nil {
			return nil, fmt.Errorf("header router: %w", err)
		}
		routers = append(routers, headerRouter)
	}

	if len(cfg.PathRoutes) > 0 {
		pathRouter, err := pathrouter.NewRouter(cfg.PathRoutes, "")
		if err != nil {
			return nil, fmt.Errorf("path router: %w", err)
		}
		routers = append(routers, pathRouter)
	}

	if cfg.DefaultUpstream != "" {
		staticRouter, err := staticrouter.NewRouter(cfg.DefaultUpstream)
		if err != nil {
			return nil, fmt.Errorf("static router: %w", err)
		}
		routers = append(routers, staticRouter)
	}

	switch len(routers) {
	case 0:
		return routingstub.NewNoOpRouter(), nil
	case 1:
		return routers[0], nil
	default:
		return chainrouter.NewRouter(routers...), nil
	}
}
