package stub

import (
	"context"

	"github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

// NoOpAuthenticator accepts every token and returns an anonymous identity.
type NoOpAuthenticator struct{}

// NewNoOpAuthenticator returns a scaffold authenticator for M0.
func NewNoOpAuthenticator() *NoOpAuthenticator {
	return &NoOpAuthenticator{}
}

// Authenticate returns an anonymous identity without validating the token.
func (a *NoOpAuthenticator) Authenticate(_ context.Context, _ string) (port.Identity, error) {
	return anonymousIdentity(), nil
}

// AuthenticateRequest returns anonymous identity without validating credentials.
func (a *NoOpAuthenticator) AuthenticateRequest(_ context.Context, _ domain.Request) (port.Identity, error) {
	return anonymousIdentity(), nil
}

func anonymousIdentity() port.Identity {
	return port.Identity{
		Subject: "anonymous",
		Claims:  map[string]string{},
	}
}
