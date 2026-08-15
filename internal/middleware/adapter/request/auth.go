package request

import (
	"context"
	"net/http"

	authctx "github.com/mauriciomendonca/universal-api-gateway/internal/auth/context"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
)

// AuthMiddleware authenticates requests via the auth port.
type AuthMiddleware struct {
	authenticator authport.RequestAuthenticator
}

// NewAuthMiddleware returns auth middleware wired to the auth port.
func NewAuthMiddleware(authenticator authport.RequestAuthenticator) *AuthMiddleware {
	return &AuthMiddleware{authenticator: authenticator}
}

// Wrap implements port.Middleware.
func (m *AuthMiddleware) Wrap(next port.Handler) port.Handler {
	return func(ctx context.Context, req domain.Request) (domain.Response, error) {
		identity, err := m.authenticator.AuthenticateRequest(ctx, req)
		if err != nil {
			return jsonResponse(http.StatusUnauthorized, "unauthorized"), nil
		}

		ctx = authctx.WithIdentity(ctx, identity)
		return next(ctx, req)
	}
}
