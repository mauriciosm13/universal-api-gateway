package header

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

type routeEntry struct {
	name     string
	value    string
	upstream string
}

// Router resolves requests by exact HTTP header match.
type Router struct {
	routes []routeEntry
}

// NewRouter returns a header router for the given routes.
func NewRouter(routes []config.HeaderRoute) (*Router, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("header router: at least one header route is required")
	}

	entries := make([]routeEntry, len(routes))
	for i, route := range routes {
		entries[i] = routeEntry{
			name:     http.CanonicalHeaderKey(route.Name),
			value:    route.Value,
			upstream: route.Upstream,
		}
	}

	return &Router{routes: entries}, nil
}

// Resolve returns the first matching header route.
func (r *Router) Resolve(_ context.Context, req domain.Request) (port.Route, error) {
	for _, route := range r.routes {
		values, ok := req.Headers[route.name]
		if !ok {
			continue
		}

		for _, value := range values {
			if value == route.value {
				return port.Route{
					ID:       fmt.Sprintf("header:%s=%s", route.name, route.value),
					Upstream: route.upstream,
				}, nil
			}
		}
	}

	return port.Route{}, port.ErrNoRoute
}
