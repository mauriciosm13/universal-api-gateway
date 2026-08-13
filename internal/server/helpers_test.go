package server

import (
	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

func newTestDependencies(cfg config.Config, router routingport.Router) Dependencies {
	if router == nil {
		router = routingstub.NewNoOpRouter()
	}

	return Dependencies{
		Config:   cfg,
		Router:   router,
		Pipeline: newTestPipeline(),
	}
}

func newTestPipeline() *requestpipeline.Pipeline {
	return requestpipeline.NewPipeline(
		requestpipeline.ContinueHandler,
		requestpipeline.NewErrorMiddleware(),
		requestpipeline.NewAuthMiddleware(authstub.NewNoOpAuthenticator()),
		requestpipeline.NewRateLimitMiddleware(ratelimitstub.NewNoOpLimiter()),
	)
}
