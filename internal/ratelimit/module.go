package ratelimit

import (
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
)

// Module registers rate limit port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	b.ProvideLimiter(ratelimitstub.NewNoOpLimiter())
}
