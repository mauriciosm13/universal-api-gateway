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
	handler     port.Handler
	capturedCtx *context.Context
}

// NewPipeline builds a pipeline from middleware and a terminal handler.
func NewPipeline(terminal port.Handler, middlewares ...port.Middleware) *Pipeline {
	var captured context.Context
	wrappingTerminal := func(ctx context.Context, req domain.Request) (domain.Response, error) {
		resp, err := terminal(ctx, req)
		captured = ctx
		return resp, err
	}

	handler := wrappingTerminal
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].Wrap(handler)
	}

	return &Pipeline{handler: handler, capturedCtx: &captured}
}

// Execute runs the middleware chain and returns enriched context.
func (p *Pipeline) Execute(ctx context.Context, req domain.Request) (context.Context, domain.Response, error) {
	*p.capturedCtx = nil

	resp, err := p.handler(ctx, req)
	if p.capturedCtx != nil && *p.capturedCtx != nil {
		return *p.capturedCtx, resp, err
	}

	return ctx, resp, err
}

// ContinueHandler signals the HTTP adapter to continue to routing and proxy.
func ContinueHandler(_ context.Context, _ domain.Request) (domain.Response, error) {
	return domain.Response{StatusCode: ContinueStatus}, nil
}
