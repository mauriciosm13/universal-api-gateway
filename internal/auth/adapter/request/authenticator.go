package request

import (
	"context"
	"errors"

	"github.com/mauriciomendonca/universal-api-gateway/internal/auth/credential"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

var errUnauthorized = errors.New("unauthorized")

// JWTAuthenticator validates Bearer tokens from requests.
type JWTAuthenticator struct {
	jwt authport.Authenticator
}

// NewJWTAuthenticator wraps a JWT token validator.
func NewJWTAuthenticator(jwt authport.Authenticator) *JWTAuthenticator {
	return &JWTAuthenticator{jwt: jwt}
}

// AuthenticateRequest validates Authorization: Bearer token.
func (a *JWTAuthenticator) AuthenticateRequest(ctx context.Context, req domain.Request) (authport.Identity, error) {
	if !credential.HasAuthorizationHeader(req) {
		return authport.Identity{}, errUnauthorized
	}

	token := credential.BearerToken(req)
	if token == "" {
		return authport.Identity{}, errUnauthorized
	}

	return a.jwt.Authenticate(ctx, token)
}

// APIKeyAuthenticator validates API keys from header or query.
type APIKeyAuthenticator struct {
	apiKey authport.Authenticator
	cfg    config.APIKeyConfig
}

// NewAPIKeyAuthenticator wraps an API key validator.
func NewAPIKeyAuthenticator(apiKey authport.Authenticator, cfg config.APIKeyConfig) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{apiKey: apiKey, cfg: cfg}
}

// AuthenticateRequest validates configured header or query API key.
func (a *APIKeyAuthenticator) AuthenticateRequest(ctx context.Context, req domain.Request) (authport.Identity, error) {
	key := credential.APIKey(req, a.cfg.HeaderName, a.cfg.QueryParam)
	if key == "" {
		return authport.Identity{}, errUnauthorized
	}

	return a.apiKey.Authenticate(ctx, key)
}
