package request

import (
	"context"
	"net/http"

	authctx "github.com/mauriciomendonca/universal-api-gateway/internal/auth/context"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
	ratelimitctx "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/context"
	ratelimitport "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/port"
)

const anonymousSubject = "anonymous"

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
		key := rateLimitKey(ctx, req)
		ctx = ratelimitctx.WithRequestPath(ctx, req.Path)

		allowed, err := m.limiter.Allow(ctx, key)
		if err != nil {
			return jsonResponse(http.StatusInternalServerError, "rate limit error"), nil
		}
		if !allowed {
			return jsonResponse(http.StatusTooManyRequests, "rate limit exceeded"), nil
		}

		return next(ctx, req)
	}
}

func rateLimitKey(ctx context.Context, req domain.Request) string {
	if id, ok := authctx.IdentityFrom(ctx); ok && id.Subject != "" && id.Subject != anonymousSubject {
		return id.Subject
	}

	return req.Host + req.Path
}
