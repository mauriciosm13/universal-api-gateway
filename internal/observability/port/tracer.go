package port

import "context"

// Tracer creates request spans for distributed tracing.
type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, func())
}
