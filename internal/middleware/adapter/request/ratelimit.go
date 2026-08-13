package request

import (
	"context"
	"net/http"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
	ratelimitport "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/port"
)

// RateLimitMiddleware enforces rate limits via the limiter port.
type RateLimitMiddleware struct {
	limiter ratelimitport.Limiter
}

// NewRateLimitMiddleware returns rate-limit middleware wired to the limiter port.
func NewRateLimitMiddleware(limiter ratelimitport.Limiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

// Wrap implements port.Middleware.
func (m *RateLimitMiddleware) Wrap(next port.Handler) port.Handler {
	return func(ctx context.Context, req domain.Request) (domain.Response, error) {
		allowed, err := m.limiter.Allow(ctx, req.Host+req.Path)
		if err != nil {
			return jsonResponse(http.StatusInternalServerError, "rate limit error"), nil
		}
		if !allowed {
			return jsonResponse(http.StatusTooManyRequests, "rate limit exceeded"), nil
		}

		return next(ctx, req)
	}
}
