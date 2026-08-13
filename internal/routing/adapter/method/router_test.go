package method

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

func TestRouterMatchesMethod(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.MethodRoute{
		{Method: "POST", Upstream: "http://post:8080"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	route, err := router.Resolve(context.Background(), domain.Request{Method: http.MethodPost})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if route.Upstream != "http://post:8080" {
		t.Fatalf("upstream = %q", route.Upstream)
	}
}

func TestRouterNoMatch(t *testing.T) {
	t.Parallel()

	router, err := NewRouter([]config.MethodRoute{
		{Method: "GET", Upstream: "http://get:8080"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	_, err = router.Resolve(context.Background(), domain.Request{Method: http.MethodPost})
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
