package auth

import (
	"fmt"

	apikeyadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/apikey"
	"github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/composite"
	jwtadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/jwt"
	requestadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/request"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
)

// Module registers auth port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	authenticator, err := buildRequestAuthenticator(b.Config())
	if err != nil {
		panic(fmt.Sprintf("auth: %v", err))
	}

	b.ProvideAuthenticator(authenticator)
}

func buildRequestAuthenticator(cfg config.Config) (authport.RequestAuthenticator, error) {
	jwtEnabled := cfg.JWT.Enabled()
	apiKeyEnabled := cfg.APIKeys.Enabled()

	switch {
	case !jwtEnabled && !apiKeyEnabled:
		return authstub.NewNoOpAuthenticator(), nil
	case jwtEnabled && !apiKeyEnabled:
		jwt, err := jwtadapter.NewValidator(cfg.JWT)
		if err != nil {
			return nil, err
		}

		return requestadapter.NewJWTAuthenticator(jwt), nil
	case !jwtEnabled && apiKeyEnabled:
		apiKey, err := apikeyadapter.NewValidator(cfg.APIKeys)
		if err != nil {
			return nil, err
		}

		return requestadapter.NewAPIKeyAuthenticator(apiKey, cfg.APIKeys), nil
	default:
		jwt, err := jwtadapter.NewValidator(cfg.JWT)
		if err != nil {
			return nil, err
		}

		apiKey, err := apikeyadapter.NewValidator(cfg.APIKeys)
		if err != nil {
			return nil, err
		}

		return composite.New(jwt, apiKey, cfg.APIKeys), nil
	}
}
