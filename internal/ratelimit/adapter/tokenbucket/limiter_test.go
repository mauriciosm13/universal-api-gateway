package tokenbucket

import (
	"context"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	ratelimitctx "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/context"
)

func TestLimiterAllowsWithinBurst(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	limiter := NewLimiterWithClock(config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 1, Burst: 3},
	}, func() time.Time { return now })

	ctx := ratelimitctx.WithRequestPath(context.Background(), "/api")

	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, "client-a")
		if err != nil {
			t.Fatalf("Allow() error = %v", err)
		}
		if !allowed {
			t.Fatalf("request %d blocked within burst", i+1)
		}
	}
}

func TestLimiterBlocksWhenEmpty(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	limiter := NewLimiterWithClock(config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 1, Burst: 2},
	}, func() time.Time { return now })

	ctx := ratelimitctx.WithRequestPath(context.Background(), "/api")

	for i := 0; i < 2; i++ {
		if allowed, _ := limiter.Allow(ctx, "client-a"); !allowed {
			t.Fatalf("request %d blocked early", i+1)
		}
	}

	allowed, err := limiter.Allow(ctx, "client-a")
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if allowed {
		t.Fatal("expected block when bucket empty")
	}
}

func TestLimiterRefillsOverTime(t *testing.T) {
	t.Parallel()

	current := time.Unix(0, 0)
	limiter := NewLimiterWithClock(config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 10, Burst: 1},
	}, func() time.Time { return current })

	ctx := ratelimitctx.WithRequestPath(context.Background(), "/api")

	if allowed, _ := limiter.Allow(ctx, "client-a"); !allowed {
		t.Fatal("expected first request allowed")
	}

	if allowed, _ := limiter.Allow(ctx, "client-a"); allowed {
		t.Fatal("expected second request blocked")
	}

	current = current.Add(200 * time.Millisecond)

	allowed, err := limiter.Allow(ctx, "client-a")
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !allowed {
		t.Fatal("expected refill after elapsed time")
	}
}

func TestLimiterRouteSpecificLimits(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	limiter := NewLimiterWithClock(config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 100, Burst: 100},
		Routes: []config.RateLimitRoute{
			{Prefix: "/api", RPS: 100, Burst: 1},
		},
	}, func() time.Time { return now })

	apiCtx := ratelimitctx.WithRequestPath(context.Background(), "/api/users")
	otherCtx := ratelimitctx.WithRequestPath(context.Background(), "/other")

	if allowed, _ := limiter.Allow(apiCtx, "client-a"); !allowed {
		t.Fatal("expected first /api request allowed")
	}
	if allowed, _ := limiter.Allow(apiCtx, "client-a"); allowed {
		t.Fatal("expected second /api request blocked")
	}

	for i := 0; i < 2; i++ {
		if allowed, _ := limiter.Allow(otherCtx, "client-a"); !allowed {
			t.Fatalf("/other request %d blocked", i+1)
		}
	}
}

func TestLimiterSeparateClients(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	limiter := NewLimiterWithClock(config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 1, Burst: 1},
	}, func() time.Time { return now })

	ctx := ratelimitctx.WithRequestPath(context.Background(), "/api")

	if allowed, _ := limiter.Allow(ctx, "client-a"); !allowed {
		t.Fatal("expected client-a first request allowed")
	}
	if allowed, _ := limiter.Allow(ctx, "client-a"); allowed {
		t.Fatal("expected client-a second request blocked")
	}

	if allowed, _ := limiter.Allow(ctx, "client-b"); !allowed {
		t.Fatal("expected client-b independent bucket")
	}
}

func TestLimiterEmptyKeyReturnsError(t *testing.T) {
	t.Parallel()

	limiter := NewLimiter(config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 1, Burst: 1},
	})

	_, err := limiter.Allow(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestLimiterCapsTokensAtBurst(t *testing.T) {
	t.Parallel()

	current := time.Unix(0, 0)
	limiter := NewLimiterWithClock(config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 1000, Burst: 2},
	}, func() time.Time { return current })

	ctx := ratelimitctx.WithRequestPath(context.Background(), "/api")

	for i := 0; i < 2; i++ {
		if allowed, _ := limiter.Allow(ctx, "client"); !allowed {
			t.Fatalf("request %d blocked within burst", i+1)
		}
	}
	if allowed, _ := limiter.Allow(ctx, "client"); allowed {
		t.Fatal("expected block at burst cap")
	}

	current = current.Add(10 * time.Second)

	for i := 0; i < 2; i++ {
		if allowed, _ := limiter.Allow(ctx, "client"); !allowed {
			t.Fatalf("refill request %d blocked", i+1)
		}
	}
	if allowed, _ := limiter.Allow(ctx, "client"); allowed {
		t.Fatal("expected block after refill burst consumed")
	}
}
