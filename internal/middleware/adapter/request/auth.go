package request

import (
	"context"
	"net/http"
	"strings"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
)

// AuthMiddleware authenticates requests via the auth port.
type AuthMiddleware struct {
	authenticator authport.Authenticator
}

// NewAuthMiddleware returns auth middleware wired to the auth port.
func NewAuthMiddleware(authenticator authport.Authenticator) *AuthMiddleware {
	return &AuthMiddleware{authenticator: authenticator}
}

// Wrap implements port.Middleware.
func (m *AuthMiddleware) Wrap(next port.Handler) port.Handler {
	return func(ctx context.Context, req domain.Request) (domain.Response, error) {
		token := bearerToken(req)
		if _, err := m.authenticator.Authenticate(ctx, token); err != nil {
			return jsonResponse(http.StatusUnauthorized, "unauthorized"), nil
		}

		return next(ctx, req)
	}
}

func bearerToken(req domain.Request) string {
	values := req.Headers["Authorization"]
	if len(values) == 0 {
		return ""
	}

	value := values[0]
	if !strings.HasPrefix(value, "Bearer ") {
		return value
	}

	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}
