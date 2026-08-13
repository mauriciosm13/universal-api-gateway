package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	pathrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/path"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

func TestGatewayProxiesToUpstream(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/hello" {
			t.Fatalf("upstream path = %q, want /api/hello", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("upstream method = %q, want GET", r.Method)
		}
		w.Header().Set("X-Upstream", "true")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "hello upstream")
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	deps := Dependencies{
		Config: config.Config{
			Host:         "127.0.0.1",
			Port:         0,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Router: router,
	}

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/hello")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	if resp.Header.Get("X-Upstream") != "true" {
		t.Fatalf("expected X-Upstream header from upstream")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if string(body) != "hello upstream" {
		t.Fatalf("body = %q, want hello upstream", string(body))
	}
}

func TestGatewayNoRouteReturns404(t *testing.T) {
	t.Parallel()

	deps := Dependencies{
		Config: config.Config{
			Host:         "127.0.0.1",
			Port:         0,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Router: routingstub.NewNoOpRouter(),
	}

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/unknown")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var body errorBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if body.Code != 404 || body.Message != "no route matched" {
		t.Fatalf("body = %+v", body)
	}
}

func TestHealthEndpointsBypassProxy(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	deps := Dependencies{
		Config: config.Config{
			Host:         "127.0.0.1",
			Port:         0,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Router: router,
	}

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	for _, path := range []string{"/health", "/health/live", "/health/ready"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("Get(%q) error = %v", path, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Get(%q) status = %d, want 200", path, resp.StatusCode)
		}
	}
}

func TestGatewayPathRouting(t *testing.T) {
	t.Parallel()

	apiUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "api:"+r.URL.Path)
	}))
	t.Cleanup(apiUpstream.Close)

	otherUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "other:"+r.URL.Path)
	}))
	t.Cleanup(otherUpstream.Close)

	router, err := pathrouter.NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: apiUpstream.URL},
	}, otherUpstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	deps := Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Router: router,
	}

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/users")
	if err != nil {
		t.Fatalf("Get(/api/users) error = %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "api:/api/users" {
		t.Fatalf("body = %q, want api:/api/users", string(body))
	}

	resp, err = http.Get(srv.URL + "/other")
	if err != nil {
		t.Fatalf("Get(/other) error = %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "other:/other" {
		t.Fatalf("body = %q, want other:/other", string(body))
	}
}

func TestToDomainRequest(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "http://gateway.local/v1/items", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Host = "gateway.local"
	req.Header.Set("X-Request-Id", "abc-123")

	domainReq := toDomainRequest(req)

	if domainReq.Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", domainReq.Method)
	}

	if domainReq.Path != "/v1/items" {
		t.Fatalf("path = %q, want /v1/items", domainReq.Path)
	}

	if domainReq.Host != "gateway.local" {
		t.Fatalf("host = %q, want gateway.local", domainReq.Host)
	}

	if got := domainReq.Headers["X-Request-Id"]; len(got) != 1 || got[0] != "abc-123" {
		t.Fatalf("headers = %v, want X-Request-Id=[abc-123]", domainReq.Headers)
	}
}
