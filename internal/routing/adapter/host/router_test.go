package host

import (
	"context"
	"errors"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

func TestRouterMatchesHost(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.HostRoute{
		{Host: "api.example.com", Upstream: "http://api:8080"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	route, err := router.Resolve(context.Background(), domain.Request{Host: "api.example.com:8080"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if route.Upstream != "http://api:8080" {
		t.Fatalf("upstream = %q", route.Upstream)
	}
}

func TestRouterNoMatch(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.HostRoute{
		{Host: "api.example.com", Upstream: "http://api:8080"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	_, err = router.Resolve(context.Background(), domain.Request{Host: "other.example.com"})
	if !errors.Is(err, routingport.ErrNoRoute) {
		t.Fatalf("Resolve() error = %v, want ErrNoRoute", err)
	}
}

func TestNewRouterRequiresRoutes(t *testing.T) {
	t.Parallel()

	if _, err := NewRouter(nil); err == nil {
		t.Fatal("NewRouter(nil) error = nil, want error")
	}
}

func TestRouterEmptyHostNoMatch(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.HostRoute{
		{Host: "api.example.com", Upstream: "http://api:8080"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	_, err = router.Resolve(context.Background(), domain.Request{})
	if !errors.Is(err, routingport.ErrNoRoute) {
		t.Fatalf("Resolve() error = %v, want ErrNoRoute", err)
	}
}
