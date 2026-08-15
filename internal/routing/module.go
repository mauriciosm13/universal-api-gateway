package routing

import (
	"fmt"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	chainrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/chain"
	headerrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/header"
	hostrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/host"
	methodrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/method"
	pathrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/path"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

// Module registers routing port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	router, err := buildRouter(b.Config())
	if err != nil {
		panic(fmt.Sprintf("routing: %v", err))
	}

	b.ProvideRouter(router)
}

func buildRouter(cfg config.Config) (routingport.Router, error) {
	routers := make([]routingport.Router, 0, 5)

	if len(cfg.HostRoutes) > 0 {
		hostRouter, err := hostrouter.NewRouter(cfg.HostRoutes)
		if err != nil {
			return nil, fmt.Errorf("host router: %w", err)
		}
		routers = append(routers, hostRouter)
	}

	if len(cfg.HeaderRoutes) > 0 {
		headerRouter, err := headerrouter.NewRouter(cfg.HeaderRoutes)
		if err != nil {
			return nil, fmt.Errorf("header router: %w", err)
		}
		routers = append(routers, headerRouter)
	}

	if len(cfg.MethodRoutes) > 0 {
		methodRouter, err := methodrouter.NewRouter(cfg.MethodRoutes)
		if err != nil {
			return nil, fmt.Errorf("method router: %w", err)
		}
		routers = append(routers, methodRouter)
	}

	if len(cfg.PathRoutes) > 0 {
		pathRouter, err := pathrouter.NewRouter(cfg.PathRoutes, nil)
		if err != nil {
			return nil, fmt.Errorf("path router: %w", err)
		}
		routers = append(routers, pathRouter)
	}

	if len(cfg.DefaultUpstreams) > 0 {
		staticRouter, err := staticrouter.NewRouter(cfg.DefaultUpstreams)
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
