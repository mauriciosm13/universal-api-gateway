package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	jwtadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/jwt"
	requestadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/request"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

const gatewayTestHMACSecret = "dev-secret-at-least-32-chars-long"

func TestGatewayJWTValidTokenProxies(t *testing.T) {
	t.Parallel()

	upstream := httptestUpstream(t, "jwt-ok")
	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	validator, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: gatewayTestHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	deps := newJWTTestDependencies(router, requestadapter.NewJWTAuthenticator(validator))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	token := signGatewayHS256Token(t, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api", nil)
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
}

func TestGatewayJWTInvalidTokenReturns401(t *testing.T) {
	t.Parallel()

	router, err := staticrouter.NewRouter(httptestUpstream(t, "unused").URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	validator, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: gatewayTestHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	deps := newJWTTestDependencies(router, requestadapter.NewJWTAuthenticator(validator))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Authorization", "Bearer not-a-valid-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestGatewayJWTMissingTokenReturns401(t *testing.T) {
	t.Parallel()

	router, err := staticrouter.NewRouter(httptestUpstream(t, "unused").URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	validator, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: gatewayTestHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	deps := newJWTTestDependencies(router, requestadapter.NewJWTAuthenticator(validator))
	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func newJWTTestDependencies(router routingport.Router, authenticator authport.RequestAuthenticator) Dependencies {
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

func signGatewayHS256Token(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(gatewayTestHMACSecret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	return signed
}

func httptestUpstream(t *testing.T, body string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
}
