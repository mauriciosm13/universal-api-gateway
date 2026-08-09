package di

import (
	"fmt"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/deps"
	middlewareport "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
	observabilityport "github.com/mauriciomendonca/universal-api-gateway/internal/observability/port"
	ratelimitport "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/port"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

// Builder assembles the gateway dependency graph from registered modules.
type Builder struct {
	cfg config.Config

	authenticator authport.Authenticator
	router        routingport.Router
	limiter       ratelimitport.Limiter
	pipeline      middlewareport.Pipeline
	logger        observabilityport.Logger
	tracer        observabilityport.Tracer
}

// NewBuilder creates a builder, registers all modules, and returns it for optional extension.
func NewBuilder(cfg config.Config, modules ...Module) *Builder {
	b := &Builder{cfg: cfg}
	for _, m := range modules {
		m.Register(b)
	}
	return b
}

// ProvideAuthenticator sets the authentication port implementation.
func (b *Builder) ProvideAuthenticator(a authport.Authenticator) {
	b.authenticator = a
}

// ProvideRouter sets the routing port implementation.
func (b *Builder) ProvideRouter(r routingport.Router) {
	b.router = r
}

// ProvideLimiter sets the rate limit port implementation.
func (b *Builder) ProvideLimiter(l ratelimitport.Limiter) {
	b.limiter = l
}

// ProvidePipeline sets the middleware pipeline port implementation.
func (b *Builder) ProvidePipeline(p middlewareport.Pipeline) {
	b.pipeline = p
}

// ProvideLogger sets the logging port implementation.
func (b *Builder) ProvideLogger(l observabilityport.Logger) {
	b.logger = l
}

// ProvideTracer sets the tracing port implementation.
func (b *Builder) ProvideTracer(t observabilityport.Tracer) {
	b.tracer = t
}

// Build validates the graph and returns gateway dependencies.
func (b *Builder) Build() (deps.Gateway, error) {
	if err := b.validate(); err != nil {
		return deps.Gateway{}, err
	}

	return deps.Gateway{
		Config:        b.cfg,
		Authenticator: b.authenticator,
		Router:        b.router,
		Limiter:       b.limiter,
		Pipeline:      b.pipeline,
		Logger:        b.logger,
		Tracer:        b.tracer,
	}, nil
}

func (b *Builder) validate() error {
	missing := make([]string, 0, 6)

	if b.authenticator == nil {
		missing = append(missing, "Authenticator")
	}
	if b.router == nil {
		missing = append(missing, "Router")
	}
	if b.limiter == nil {
		missing = append(missing, "Limiter")
	}
	if b.pipeline == nil {
		missing = append(missing, "Pipeline")
	}
	if b.logger == nil {
		missing = append(missing, "Logger")
	}
	if b.tracer == nil {
		missing = append(missing, "Tracer")
	}

	if len(missing) > 0 {
		return fmt.Errorf("di: missing providers: %v", missing)
	}

	return nil
}
