package stub

import "context"

// NoOpLimiter is a scaffold limiter that allows all requests.
type NoOpLimiter struct{}

// NewNoOpLimiter returns an allow-all limiter for M0.
func NewNoOpLimiter() *NoOpLimiter {
	return &NoOpLimiter{}
}

// Allow always permits the request.
func (l *NoOpLimiter) Allow(_ context.Context, _ string) (bool, error) {
	return true, nil
}
