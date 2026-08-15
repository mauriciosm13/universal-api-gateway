package composite

import (
	"context"
	"errors"

	"github.com/mauriciomendonca/universal-api-gateway/internal/auth/credential"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

var errUnauthorized = errors.New("unauthorized")

// Authenticator tries JWT Bearer first, then API key credentials.
type Authenticator struct {
	jwt    authport.Authenticator
	apiKey authport.Authenticator
	cfg    config.APIKeyConfig
}

// New builds a composite request authenticator.
func New(jwt, apiKey authport.Authenticator, cfg config.APIKeyConfig) *Authenticator {
	return &Authenticator{jwt: jwt, apiKey: apiKey, cfg: cfg}
}

// AuthenticateRequest validates JWT or API key credentials from req.
func (a *Authenticator) AuthenticateRequest(ctx context.Context, req domain.Request) (authport.Identity, error) {
	if credential.HasAuthorizationHeader(req) {
		token := credential.BearerToken(req)
		if token == "" {
			return authport.Identity{}, errUnauthorized
		}

		return a.jwt.Authenticate(ctx, token)
	}

	key := credential.APIKey(req, a.cfg.HeaderName, a.cfg.QueryParam)
	if key == "" {
		return authport.Identity{}, errUnauthorized
	}

	return a.apiKey.Authenticate(ctx, key)
}
