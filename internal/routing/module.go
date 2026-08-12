package routing

import (
	"fmt"

	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	pathrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/path"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
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
		panic(fmt.Sprintf("routing: static router: %v", err))
	}

	b.ProvideRouter(router)
}
