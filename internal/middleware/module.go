package middleware

import (
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
)

// Module registers middleware port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	pipeline := requestpipeline.NewPipeline(
		requestpipeline.ContinueHandler,
		requestpipeline.NewErrorMiddleware(),
		requestpipeline.NewAuthMiddleware(b.Authenticator()),
		requestpipeline.NewRateLimitMiddleware(b.Limiter()),
	)

	b.ProvidePipeline(pipeline)
}
