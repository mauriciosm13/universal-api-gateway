package deps

import (
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	middlewareport "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
	observabilityport "github.com/mauriciomendonca/universal-api-gateway/internal/observability/port"
	ratelimitport "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/port"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

// Gateway holds injected gateway ports and configuration.
type Gateway struct {
	Config        config.Config
	Authenticator authport.Authenticator
	Router        routingport.Router
	Limiter       ratelimitport.Limiter
	Pipeline      middlewareport.Pipeline
	Logger        observabilityport.Logger
	Tracer        observabilityport.Tracer
}
