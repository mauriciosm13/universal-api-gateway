package ratelimit

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	tokenbucket "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/adapter/tokenbucket"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
)

func TestBuildLimiterNoOpWhenDisabled(t *testing.T) {
	t.Parallel()

	limiter, err := buildLimiter(config.Config{})
	if err != nil {
		t.Fatalf("buildLimiter() error = %v", err)
	}

	if _, ok := limiter.(*ratelimitstub.NoOpLimiter); !ok {
		t.Fatalf("limiter type = %T, want *stub.NoOpLimiter", limiter)
	}
}

func TestBuildLimiterTokenBucketWhenEnabled(t *testing.T) {
	t.Parallel()

	limiter, err := buildLimiter(config.Config{
		RateLimit: config.RateLimitConfig{
			Global: &config.RateLimitParams{RPS: 1, Burst: 1},
		},
	})
	if err != nil {
		t.Fatalf("buildLimiter() error = %v", err)
	}

	if _, ok := limiter.(*tokenbucket.Limiter); !ok {
		t.Fatalf("limiter type = %T, want *tokenbucket.Limiter", limiter)
	}
}
