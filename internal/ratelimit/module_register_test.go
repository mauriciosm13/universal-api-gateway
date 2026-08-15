package ratelimit

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	tokenbucket "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/adapter/tokenbucket"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
)

func TestModuleRegisterProvidesLimiter(t *testing.T) {
	t.Parallel()

	builder := di.NewBuilder(config.Config{
		RateLimit: config.RateLimitConfig{
			Global: &config.RateLimitParams{RPS: 1, Burst: 1},
		},
	}, Module{})

	if builder.Limiter() == nil {
		t.Fatal("expected limiter to be registered")
	}

	if _, ok := builder.Limiter().(*tokenbucket.Limiter); !ok {
		t.Fatalf("limiter type = %T, want *tokenbucket.Limiter", builder.Limiter())
	}
}

func TestModuleRegisterProvidesNoOpWhenDisabled(t *testing.T) {
	t.Parallel()

	builder := di.NewBuilder(config.Config{}, Module{})

	if _, ok := builder.Limiter().(*ratelimitstub.NoOpLimiter); !ok {
		t.Fatalf("limiter type = %T, want *stub.NoOpLimiter", builder.Limiter())
	}
}
