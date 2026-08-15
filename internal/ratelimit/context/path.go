package context

import "context"

type requestPathKey struct{}

// WithRequestPath stores the request path on ctx for route-aware rate limiting.
func WithRequestPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, requestPathKey{}, path)
}

// RequestPathFrom returns the path stored for rate limiting, if any.
func RequestPathFrom(ctx context.Context) (string, bool) {
	path, ok := ctx.Value(requestPathKey{}).(string)
	return path, ok
}
