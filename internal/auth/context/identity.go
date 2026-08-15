package context

import (
	"context"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
)

type identityKey struct{}

// WithIdentity stores authenticated identity on ctx.
func WithIdentity(ctx context.Context, id authport.Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

// IdentityFrom returns identity stored by auth middleware, if any.
func IdentityFrom(ctx context.Context) (authport.Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(authport.Identity)
	return id, ok
}
