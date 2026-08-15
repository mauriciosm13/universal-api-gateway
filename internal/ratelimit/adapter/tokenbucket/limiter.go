package tokenbucket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	ratelimitctx "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/context"
)

const bucketKeySeparator = "\x00"

// Limiter implements token bucket rate limiting in memory.
// Buckets are not shared across gateway instances.
type Limiter struct {
	cfg   config.RateLimitConfig
	now   func() time.Time
	store sync.Map
}

type bucketState struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
	rate   float64
	burst  float64
}

// NewLimiter returns a token bucket limiter for cfg.
func NewLimiter(cfg config.RateLimitConfig) *Limiter {
	return &Limiter{
		cfg: cfg,
		now: time.Now,
	}
}

// NewLimiterWithClock returns a limiter that uses now for token refill (tests).
func NewLimiterWithClock(cfg config.RateLimitConfig, now func() time.Time) *Limiter {
	return &Limiter{
		cfg: cfg,
		now: now,
	}
}

// Allow consumes one token for key. Returns (false, nil) when the limit is exceeded.
func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("rate limit key is required")
	}

	path, ok := ratelimitctx.RequestPathFrom(ctx)
	if !ok {
		path = "/"
	}

	params, profile := l.cfg.LimitsForPath(path)
	bucketKey := key + bucketKeySeparator + profile

	allowed := l.allowBucket(bucketKey, params.RPS, params.Burst)
	return allowed, nil
}

func (l *Limiter) allowBucket(bucketKey string, rate float64, burst int) bool {
	now := l.now()

	stateAny, _ := l.store.LoadOrStore(bucketKey, &bucketState{
		tokens: float64(burst),
		last:   now,
		rate:   rate,
		burst:  float64(burst),
	})
	state := stateAny.(*bucketState)

	state.mu.Lock()
	defer state.mu.Unlock()

	state.rate = rate
	state.burst = float64(burst)

	elapsed := now.Sub(state.last).Seconds()
	if elapsed > 0 {
		state.tokens = min(state.burst, state.tokens+elapsed*state.rate)
		state.last = now
	}

	if state.tokens >= 1 {
		state.tokens -= 1
		return true
	}

	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
