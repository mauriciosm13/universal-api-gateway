package static

import (
	"context"
	"fmt"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

const defaultRouteID = "default"

// Router resolves every request to a single configured upstream.
type Router struct {
	upstream string
}

// NewRouter returns a static router for the given upstream URL.
func NewRouter(upstream string) (*Router, error) {
	if err := config.ValidateUpstreamURL(upstream); err != nil {
		return nil, fmt.Errorf("static router: %w", err)
	}

	return &Router{upstream: upstream}, nil
}

// Resolve returns the default route for any request.
func (r *Router) Resolve(_ context.Context, _ domain.Request) (port.Route, error) {
	return port.Route{
		ID:       defaultRouteID,
		Upstream: r.upstream,
	}, nil
}
