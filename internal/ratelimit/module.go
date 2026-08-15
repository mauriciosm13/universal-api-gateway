package ratelimit

import (
	"fmt"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	tokenbucket "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/adapter/tokenbucket"
	ratelimitport "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/port"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
)

// Module registers rate limit port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	limiter, err := buildLimiter(b.Config())
	if err != nil {
		panic(fmt.Sprintf("ratelimit: %v", err))
	}

	b.ProvideLimiter(limiter)
}

func buildLimiter(cfg config.Config) (ratelimitport.Limiter, error) {
	if !cfg.RateLimit.Enabled() {
		return ratelimitstub.NewNoOpLimiter(), nil
	}

	return tokenbucket.NewLimiter(cfg.RateLimit), nil
}
