package header

import (
	"context"
	"errors"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

func TestRouterMatchesHeader(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.HeaderRoute{
		{Name: "X-Version", Value: "v1", Upstream: "http://v1:8080"},
		{Name: "X-Version", Value: "v2", Upstream: "http://v2:8080"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	route, err := router.Resolve(context.Background(), domain.Request{
		Headers: map[string][]string{"X-Version": {"v2"}},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if route.Upstream != "http://v2:8080" {
		t.Fatalf("upstream = %q, want http://v2:8080", route.Upstream)
	}
}

func TestRouterCanonicalHeaderName(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.HeaderRoute{
		{Name: "x-version", Value: "v1", Upstream: "http://v1:8080"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	route, err := router.Resolve(context.Background(), domain.Request{
		Headers: map[string][]string{"X-Version": {"v1"}},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if route.Upstream != "http://v1:8080" {
		t.Fatalf("upstream = %q, want http://v1:8080", route.Upstream)
	}
}

func TestRouterNoMatch(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.HeaderRoute{
		{Name: "X-Version", Value: "v1", Upstream: "http://v1:8080"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	_, err = router.Resolve(context.Background(), domain.Request{
		Headers: map[string][]string{"X-Version": {"v3"}},
	})
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
