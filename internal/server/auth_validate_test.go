package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	jwtadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/jwt"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
)

func TestAuthValidateNoOpReturnsAnonymous(t *testing.T) {
	t.Parallel()

	deps := authValidateDependencies(authstub.NewNoOpAuthenticator())
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/auth/validate")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body authValidateBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if body.Subject != "anonymous" {
		t.Fatalf("subject = %q, want anonymous", body.Subject)
	}
}

func TestAuthValidateJWTValidToken(t *testing.T) {
	t.Parallel()

	validator, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: gatewayTestHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	deps := authValidateDependencies(validator)
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	token := signGatewayHS256Token(t, jwt.MapClaims{
		"sub":  "user-123",
		"role": "admin",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/auth/validate", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body authValidateBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if body.Subject != "user-123" || body.Claims["role"] != "admin" {
		t.Fatalf("body = %+v", body)
	}
}

func TestAuthValidateJWTMissingTokenReturns401(t *testing.T) {
	t.Parallel()

	validator, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: gatewayTestHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	deps := authValidateDependencies(validator)
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/auth/validate")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestAuthValidateJWTInvalidTokenReturns401(t *testing.T) {
	t.Parallel()

	validator, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: gatewayTestHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	deps := authValidateDependencies(validator)
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/auth/validate", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Authorization", "Bearer not-valid")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func authValidateDependencies(authenticator authport.Authenticator) Dependencies {
	return Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Authenticator: authenticator,
		Pipeline: requestpipeline.NewPipeline(
			requestpipeline.ContinueHandler,
			requestpipeline.NewErrorMiddleware(),
			requestpipeline.NewAuthMiddleware(authenticator),
			requestpipeline.NewRateLimitMiddleware(ratelimitstub.NewNoOpLimiter()),
		),
	}
}
