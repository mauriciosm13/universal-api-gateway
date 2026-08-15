package port

import (
	"context"
	"errors"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

// ErrNoRoute indicates that no route matched the request.
var ErrNoRoute = errors.New("routing: no route matched")

// Route describes upstream targets for a matched request.
type Route struct {
	ID        string
	Upstreams []string
}

// Router resolves a domain request to a route.
type Router interface {
	Resolve(ctx context.Context, req domain.Request) (Route, error)
}
