package routing

import (
	"fmt"

	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

// Module registers routing port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	upstream := b.Config().DefaultUpstream
	if upstream == "" {
		b.ProvideRouter(routingstub.NewNoOpRouter())
		return
	}

	router, err := staticrouter.NewRouter(upstream)
	if err != nil {
		panic(fmt.Sprintf("routing: static router: %v", err))
	}

	b.ProvideRouter(router)
}
