package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	apikeyadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/apikey"
	"github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/composite"
	jwtadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/jwt"
	requestadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/request"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
)

const apiKeyTestSecret = "dev-secret-at-least-32-chars-long"

func TestGatewayAPIKeyValidHeaderProxiesWithUserID(t *testing.T) {
	t.Parallel()

	var upstreamUserID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamUserID = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(upstream.Close)

	deps := newAPIKeyTestDependencies(t, upstream.URL, testAPIKeyAuthenticator(t, false))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-API-Key", "dev-key-abc")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	if upstreamUserID != "local-dev" {
		t.Fatalf("upstream X-User-Id = %q, want local-dev", upstreamUserID)
	}
}

func TestGatewayAPIKeyValidQueryProxies(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	deps := newAPIKeyTestDependencies(t, upstream.URL, testAPIKeyAuthenticator(t, false))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api?api_key=dev-key-abc")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestGatewayAPIKeyInvalidReturns401(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("upstream should not be called")
	}))
	t.Cleanup(upstream.Close)

	deps := newAPIKeyTestDependencies(t, upstream.URL, testAPIKeyAuthenticator(t, false))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-API-Key", "wrong-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestGatewayCompositeAcceptsJWTOrAPIKey(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	deps := newAPIKeyTestDependencies(t, upstream.URL, testAPIKeyAuthenticator(t, true))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	token := signGatewayHS256Token(t, jwt.MapClaims{
		"sub": "jwt-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	jwtReq, err := http.NewRequest(http.MethodGet, srv.URL+"/api", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	jwtReq.Header.Set("Authorization", "Bearer "+token)

	jwtResp, err := http.DefaultClient.Do(jwtReq)
	if err != nil {
		t.Fatalf("Do(JWT) error = %v", err)
	}
	jwtResp.Body.Close()
	if jwtResp.StatusCode != http.StatusOK {
		t.Fatalf("JWT status = %d, want 200", jwtResp.StatusCode)
	}

	keyReq, err := http.NewRequest(http.MethodGet, srv.URL+"/api", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	keyReq.Header.Set("X-API-Key", "dev-key-abc")

	keyResp, err := http.DefaultClient.Do(keyReq)
	if err != nil {
		t.Fatalf("Do(API key) error = %v", err)
	}
	keyResp.Body.Close()
	if keyResp.StatusCode != http.StatusOK {
		t.Fatalf("API key status = %d, want 200", keyResp.StatusCode)
	}
}

func TestGatewayDoesNotOverwriteClientUserIDHeader(t *testing.T) {
	t.Parallel()

	var upstreamUserID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamUserID = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	deps := newAPIKeyTestDependencies(t, upstream.URL, testAPIKeyAuthenticator(t, false))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-API-Key", "dev-key-abc")
	req.Header.Set("X-User-Id", "client-user")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if upstreamUserID != "client-user" {
		t.Fatalf("upstream X-User-Id = %q, want client-user", upstreamUserID)
	}
}

func TestAuthValidateAPIKeyReturnsIdentity(t *testing.T) {
	t.Parallel()

	deps := authValidateDependencies(testAPIKeyAuthenticator(t, false))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/auth/validate", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-API-Key", "dev-key-abc")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func testAPIKeyAuthenticator(t *testing.T, withJWT bool) authport.RequestAuthenticator {
	t.Helper()

	apiKeyCfg := config.APIKeyConfig{
		Keys:       map[string]string{"dev-key-abc": "local-dev"},
		HeaderName: "X-API-Key",
		QueryParam: "api_key",
	}

	apiKey, err := apikeyadapter.NewValidator(apiKeyCfg)
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	if !withJWT {
		return requestadapter.NewAPIKeyAuthenticator(apiKey, apiKeyCfg)
	}

	jwt, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: apiKeyTestSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	return composite.New(jwt, apiKey, apiKeyCfg)
}

func newAPIKeyTestDependencies(t *testing.T, upstreamURL string, authenticator authport.RequestAuthenticator) Dependencies {
	t.Helper()

	router, err := staticrouter.NewRouter(upstreamURL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	return Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Router: router,
		Pipeline: requestpipeline.NewPipeline(
			requestpipeline.ContinueHandler,
			requestpipeline.NewErrorMiddleware(),
			requestpipeline.NewAuthMiddleware(authenticator),
			requestpipeline.NewRateLimitMiddleware(ratelimitstub.NewNoOpLimiter()),
		),
	}
}
