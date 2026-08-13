package path

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

const defaultRouteID = "default"

type routeEntry struct {
	prefix   string
	upstream string
}

// Router resolves requests using longest prefix match.
type Router struct {
	routes          []routeEntry
	defaultUpstream string
}

// NewRouter returns a path router for the given routes and optional default upstream.
func NewRouter(routes []config.PathRoute, defaultUpstream string) (*Router, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("path router: at least one path route is required")
	}

	if defaultUpstream != "" {
		if err := config.ValidateUpstreamURL(defaultUpstream); err != nil {
			return nil, fmt.Errorf("path router: default upstream: %w", err)
		}
	}

	entries := make([]routeEntry, len(routes))
	for i, route := range routes {
		entries[i] = routeEntry{
			prefix:   route.Prefix,
			upstream: route.Upstream,
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if len(entries[i].prefix) != len(entries[j].prefix) {
			return len(entries[i].prefix) > len(entries[j].prefix)
		}
		return entries[i].prefix < entries[j].prefix
	})

	return &Router{
		routes:          entries,
		defaultUpstream: defaultUpstream,
	}, nil
}

// Resolve returns the longest matching prefix route or the default upstream.
func (r *Router) Resolve(_ context.Context, req domain.Request) (port.Route, error) {
	for _, route := range r.routes {
		if matchesPrefix(req.Path, route.prefix) {
			return port.Route{
				ID:       route.prefix,
				Upstream: route.upstream,
			}, nil
		}
	}

	if r.defaultUpstream != "" {
		return port.Route{
			ID:       defaultRouteID,
			Upstream: r.defaultUpstream,
		}, nil
	}

	return port.Route{}, port.ErrNoRoute
}

func matchesPrefix(path, prefix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}

	if len(path) == len(prefix) {
		return true
	}

	return path[len(prefix)] == '/'
}
