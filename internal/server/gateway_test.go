package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
	chainrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/chain"
	headerrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/header"
	hostrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/host"
	methodrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/method"
	pathrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/path"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
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

	deps := newTestDependencies(config.Config{
		Host:         "127.0.0.1",
		Port:         0,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, router)

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

	deps := newTestDependencies(config.Config{
		Host:         "127.0.0.1",
		Port:         0,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, routingstub.NewNoOpRouter())

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

	deps := newTestDependencies(config.Config{
		Host:         "127.0.0.1",
		Port:         0,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, router)

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

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, router)

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

func TestGatewayHeaderRouting(t *testing.T) {
	t.Parallel()

	v1Upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "v1")
	}))
	t.Cleanup(v1Upstream.Close)

	pathUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "path")
	}))
	t.Cleanup(pathUpstream.Close)

	headerRouter, err := headerrouter.NewRouter([]config.HeaderRoute{
		{Name: "X-Version", Value: "v1", Upstream: v1Upstream.URL},
	})
	if err != nil {
		t.Fatalf("header NewRouter() error = %v", err)
	}

	pathRouter, err := pathrouter.NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: pathUpstream.URL},
	}, "")
	if err != nil {
		t.Fatalf("path NewRouter() error = %v", err)
	}

	router := chainrouter.NewRouter(headerRouter, pathRouter)

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, router)

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/users", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-Version", "v1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "v1" {
		t.Fatalf("header match body = %q, want v1", string(body))
	}

	resp, err = http.Get(srv.URL + "/api/users")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "path" {
		t.Fatalf("path match body = %q, want path", string(body))
	}
}

func TestGatewayHostRouting(t *testing.T) {
	t.Parallel()

	hostUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "host-upstream")
	}))
	t.Cleanup(hostUpstream.Close)

	pathUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "path-upstream")
	}))
	t.Cleanup(pathUpstream.Close)

	hostRouter, err := hostrouter.NewRouter([]config.HostRoute{
		{Host: "api.example.com", Upstream: hostUpstream.URL},
	})
	if err != nil {
		t.Fatalf("host NewRouter() error = %v", err)
	}

	pathRouter, err := pathrouter.NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: pathUpstream.URL},
	}, "")
	if err != nil {
		t.Fatalf("path NewRouter() error = %v", err)
	}

	router := chainrouter.NewRouter(hostRouter, pathRouter)
	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, router)

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/users", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Host = "api.example.com"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "host-upstream" {
		t.Fatalf("body = %q, want host-upstream", string(body))
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

func TestGatewayPipelineError(t *testing.T) {
	t.Parallel()

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, routingstub.NewNoOpRouter())
	deps.Pipeline = errorPipeline{}

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusInternalServerError, "pipeline error")
}

func TestGatewayRoutingError(t *testing.T) {
	t.Parallel()

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, errorRouter{})

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusInternalServerError, "routing error")
}

func TestGatewayInvalidUpstream(t *testing.T) {
	t.Parallel()

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, badUpstreamRouter{})

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusInternalServerError, "invalid upstream")
}

func TestGatewayRateLimitBlocked(t *testing.T) {
	t.Parallel()

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, routingstub.NewNoOpRouter())
	deps.Pipeline = requestpipeline.NewPipeline(
		requestpipeline.ContinueHandler,
		requestpipeline.NewErrorMiddleware(),
		requestpipeline.NewRateLimitMiddleware(blockGatewayLimiter{}),
	)

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusTooManyRequests, "rate limit exceeded")
}

func TestGatewayMethodRouting(t *testing.T) {
	t.Parallel()

	postUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "post-upstream")
	}))
	t.Cleanup(postUpstream.Close)

	getUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "get-upstream")
	}))
	t.Cleanup(getUpstream.Close)

	methodRouter, err := methodrouter.NewRouter([]config.MethodRoute{
		{Method: "POST", Upstream: postUpstream.URL},
	})
	if err != nil {
		t.Fatalf("method NewRouter() error = %v", err)
	}

	pathRouter, err := pathrouter.NewRouter([]config.PathRoute{
		{Prefix: "/api", Upstream: getUpstream.URL},
	}, "")
	if err != nil {
		t.Fatalf("path NewRouter() error = %v", err)
	}

	router := chainrouter.NewRouter(methodRouter, pathRouter)
	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, router)

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/items", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do(POST) error = %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "post-upstream" {
		t.Fatalf("POST body = %q, want post-upstream", string(body))
	}

	resp, err = http.Get(srv.URL + "/api/items")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "get-upstream" {
		t.Fatalf("GET body = %q, want get-upstream", string(body))
	}
}

type errorPipeline struct{}

func (errorPipeline) Execute(_ context.Context, _ domain.Request) (domain.Response, error) {
	return domain.Response{}, errors.New("pipeline failed")
}

type errorRouter struct{}

func (errorRouter) Resolve(_ context.Context, _ domain.Request) (routingport.Route, error) {
	return routingport.Route{}, errors.New("routing failed")
}

type badUpstreamRouter struct{}

func (badUpstreamRouter) Resolve(_ context.Context, _ domain.Request) (routingport.Route, error) {
	return routingport.Route{Upstream: "://invalid"}, nil
}

type blockGatewayLimiter struct{}

func (blockGatewayLimiter) Allow(context.Context, string) (bool, error) {
	return false, nil
}

func assertJSONError(t *testing.T, resp *http.Response, wantStatus int, wantMessage string) {
	t.Helper()

	if resp.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", resp.StatusCode, wantStatus)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var body errorBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if body.Code != wantStatus || body.Message != wantMessage {
		t.Fatalf("body = %+v, want code=%d message=%q", body, wantStatus, wantMessage)
	}
}
