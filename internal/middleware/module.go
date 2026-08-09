package middleware

import (
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	middlewarestub "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/stub"
)

// Module registers middleware port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	b.ProvidePipeline(middlewarestub.NewPassthroughPipeline())
}
