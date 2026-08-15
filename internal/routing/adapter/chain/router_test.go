package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

type stubRouter struct {
	route routingport.Route
	err   error
}

func (s stubRouter) Resolve(context.Context, domain.Request) (routingport.Route, error) {
	return s.route, s.err
}

func TestRouterUsesFirstMatch(t *testing.T) {
	t.Parallel()

	first := stubRouter{err: routingport.ErrNoRoute}
	second := stubRouter{
		route: routingport.Route{ID: "second", Upstreams: []string{"http://second:8080"}},
	}

	router := NewRouter(first, second)

	route, err := router.Resolve(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if route.ID != "second" {
		t.Fatalf("route ID = %q, want second", route.ID)
	}
}

func TestRouterReturnsNoRoute(t *testing.T) {
	t.Parallel()

	router := NewRouter(
		stubRouter{err: routingport.ErrNoRoute},
		stubRouter{err: routingport.ErrNoRoute},
	)

	_, err := router.Resolve(context.Background(), domain.Request{})
	if !errors.Is(err, routingport.ErrNoRoute) {
		t.Fatalf("Resolve() error = %v, want ErrNoRoute", err)
	}
}

func TestRouterPropagatesError(t *testing.T) {
	t.Parallel()

	router := NewRouter(stubRouter{err: errors.New("boom")})

	_, err := router.Resolve(context.Background(), domain.Request{})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("Resolve() error = %v, want boom", err)
	}
}
