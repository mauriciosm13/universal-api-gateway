package path

import (
	"context"
	"errors"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

func TestRouterLongestPrefixMatch(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: "http://api:8080"},
		{Prefix: "/api/v2", Upstream: "http://api-v2:8080"},
	}, "")
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	tests := []struct {
		path     string
		wantID   string
		wantHost string
	}{
		{path: "/api/v2/users", wantID: "/api/v2", wantHost: "http://api-v2:8080"},
		{path: "/api/users", wantID: "/api", wantHost: "http://api:8080"},
		{path: "/api", wantID: "/api", wantHost: "http://api:8080"},
	}

	for _, tt := range tests {
		route, err := router.Resolve(context.Background(), domain.Request{Path: tt.path})
		if err != nil {
			t.Fatalf("Resolve(%q) error = %v", tt.path, err)
		}

		if route.ID != tt.wantID {
			t.Fatalf("Resolve(%q) ID = %q, want %q", tt.path, route.ID, tt.wantID)
		}

		if route.Upstream != tt.wantHost {
			t.Fatalf("Resolve(%q) upstream = %q, want %q", tt.path, route.Upstream, tt.wantHost)
		}
	}
}

func TestRouterPrefixBoundary(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: "http://api:8080"},
	}, "")
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	_, err = router.Resolve(context.Background(), domain.Request{Path: "/apiv2"})
	if !errors.Is(err, routingport.ErrNoRoute) {
		t.Fatalf("Resolve(/apiv2) error = %v, want ErrNoRoute", err)
	}
}

func TestRouterDefaultFallback(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: "http://api:8080"},
	}, "http://default:8080")
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	route, err := router.Resolve(context.Background(), domain.Request{Path: "/other"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if route.ID != "default" || route.Upstream != "http://default:8080" {
		t.Fatalf("route = %+v, want default upstream", route)
	}
}

func TestRouterNoMatchWithoutDefault(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: "http://api:8080"},
	}, "")
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	_, err = router.Resolve(context.Background(), domain.Request{Path: "/other"})
	if !errors.Is(err, routingport.ErrNoRoute) {
		t.Fatalf("Resolve() error = %v, want ErrNoRoute", err)
	}
}

func TestNewRouterRequiresRoutes(t *testing.T) {
	t.Parallel()

	if _, err := NewRouter(nil, "http://default:8080"); err == nil {
		t.Fatal("NewRouter(nil) error = nil, want error")
	}
}

func TestNewRouterValidatesDefaultUpstream(t *testing.T) {
	t.Parallel()

	if _, err := NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: "http://api:8080"},
	}, "not-a-url"); err == nil {
		t.Fatal("NewRouter() error = nil, want error")
	}
}
