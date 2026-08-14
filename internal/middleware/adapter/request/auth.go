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
		token := BearerToken(req)
		if _, err := m.authenticator.Authenticate(ctx, token); err != nil {
			return jsonResponse(http.StatusUnauthorized, "unauthorized"), nil
		}

		return next(ctx, req)
	}
}

// BearerToken extracts the token from Authorization: Bearer <token>.
// Malformed or missing Bearer prefix returns an empty string.
func BearerToken(req domain.Request) string {
	values := req.Headers["Authorization"]
	if len(values) == 0 {
		return ""
	}

	value := strings.TrimSpace(values[0])
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}
