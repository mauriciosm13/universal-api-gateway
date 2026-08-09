package routing

import (
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

// Module registers routing port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	b.ProvideRouter(routingstub.NewNoOpRouter())
}
