package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
	tokenbucket "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/adapter/tokenbucket"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
)

func newRateLimitTestDependencies(t *testing.T, cfg config.RateLimitConfig) Dependencies {
	t.Helper()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	return Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
			RateLimit:    cfg,
		},
		Router: router,
		Pipeline: requestpipeline.NewPipeline(
			requestpipeline.ContinueHandler,
			requestpipeline.NewErrorMiddleware(),
			requestpipeline.NewAuthMiddleware(authstub.NewNoOpAuthenticator()),
			requestpipeline.NewRateLimitMiddleware(tokenbucket.NewLimiter(cfg)),
		),
	}
}

func TestGatewayRateLimitBurstThen429(t *testing.T) {
	t.Parallel()

	deps := newRateLimitTestDependencies(t, config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 100, Burst: 3},
	})

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	for i := 0; i < 3; i++ {
		resp, err := http.Get(srv.URL + "/api")
		if err != nil {
			t.Fatalf("Get() %d error = %v", i+1, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, resp.StatusCode)
		}
	}

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusTooManyRequests, "rate limit exceeded")
}

func TestGatewayRateLimitRefillAllowsTraffic(t *testing.T) {
	t.Parallel()

	current := time.Unix(0, 0)
	cfg := config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 10, Burst: 1},
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	deps := Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
			RateLimit:    cfg,
		},
		Router: router,
		Pipeline: requestpipeline.NewPipeline(
			requestpipeline.ContinueHandler,
			requestpipeline.NewErrorMiddleware(),
			requestpipeline.NewAuthMiddleware(authstub.NewNoOpAuthenticator()),
			requestpipeline.NewRateLimitMiddleware(tokenbucket.NewLimiterWithClock(cfg, func() time.Time {
				return current
			})),
		),
	}

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first status = %d, want 200", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want 429", resp.StatusCode)
	}

	current = current.Add(200 * time.Millisecond)

	resp, err = http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("third status = %d, want 200 after refill", resp.StatusCode)
	}
}

func TestGatewayRateLimitRouteOverrideStricterThanGlobal(t *testing.T) {
	t.Parallel()

	deps := newRateLimitTestDependencies(t, config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 100, Burst: 100},
		Routes: []config.RateLimitRoute{
			{Prefix: "/api", RPS: 100, Burst: 1},
		},
	})

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() /api error = %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api first status = %d, want 200", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() /api error = %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("/api second status = %d, want 429", resp.StatusCode)
	}

	for i := 0; i < 2; i++ {
		resp, err = http.Get(srv.URL + "/other")
		if err != nil {
			t.Fatalf("Get() /other error = %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("/other request %d status = %d, want 200", i+1, resp.StatusCode)
		}
	}
}

func TestGatewayRateLimitBlockedDoesNotCallUpstream(t *testing.T) {
	t.Parallel()

	called := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	cfg := config.RateLimitConfig{
		Global: &config.RateLimitParams{RPS: 1, Burst: 0},
	}

	deps := Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
			RateLimit:    cfg,
		},
		Router: router,
		Pipeline: requestpipeline.NewPipeline(
			requestpipeline.ContinueHandler,
			requestpipeline.NewErrorMiddleware(),
			requestpipeline.NewAuthMiddleware(authstub.NewNoOpAuthenticator()),
			requestpipeline.NewRateLimitMiddleware(tokenbucket.NewLimiter(cfg)),
		),
	}

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusTooManyRequests, "rate limit exceeded")
	if called {
		t.Fatal("upstream called when rate limited")
	}
}
