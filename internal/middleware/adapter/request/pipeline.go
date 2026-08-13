package request

import (
	"context"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
)

// ContinueStatus means the pipeline finished and the HTTP handler should proxy.
const ContinueStatus = 0

// Pipeline executes composed middleware before reverse proxy handling.
type Pipeline struct {
	handler port.Handler
}

// NewPipeline builds a pipeline from middleware and a terminal handler.
func NewPipeline(terminal port.Handler, middlewares ...port.Middleware) *Pipeline {
	handler := terminal
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].Wrap(handler)
	}

	return &Pipeline{handler: handler}
}

// Execute runs the middleware chain.
func (p *Pipeline) Execute(ctx context.Context, req domain.Request) (domain.Response, error) {
	return p.handler(ctx, req)
}

// ContinueHandler signals the HTTP adapter to continue to routing and proxy.
func ContinueHandler(_ context.Context, _ domain.Request) (domain.Response, error) {
	return domain.Response{StatusCode: ContinueStatus}, nil
}
