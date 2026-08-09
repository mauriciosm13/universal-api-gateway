package port

import "context"

// Limiter decides whether a request key may proceed.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
