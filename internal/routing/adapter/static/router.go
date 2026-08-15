package static

import (
	"context"
	"fmt"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

const defaultRouteID = "default"

// Router resolves every request to configured upstreams.
type Router struct {
	upstreams []string
}

// NewRouter returns a static router for the given upstream URLs.
func NewRouter(upstreams []string) (*Router, error) {
	if len(upstreams) == 0 {
		return nil, fmt.Errorf("static router: at least one upstream is required")
	}

	for _, upstream := range upstreams {
		if err := config.ValidateUpstreamURL(upstream); err != nil {
			return nil, fmt.Errorf("static router: %w", err)
		}
	}

	return &Router{upstreams: upstreams}, nil
}

// Resolve returns the default route for any request.
func (r *Router) Resolve(_ context.Context, _ domain.Request) (port.Route, error) {
	return port.Route{
		ID:        defaultRouteID,
		Upstreams: r.upstreams,
	}, nil
}
