package port

import (
	"context"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

// Handler processes a gateway request and returns a response.
type Handler func(ctx context.Context, req domain.Request) (domain.Response, error)

// Middleware wraps a handler in the request pipeline.
type Middleware interface {
	Wrap(next Handler) Handler
}

// Pipeline executes the composed middleware chain.
type Pipeline interface {
	Execute(ctx context.Context, req domain.Request) (domain.Response, error)
}
