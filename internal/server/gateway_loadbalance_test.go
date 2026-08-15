package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
)

func TestGatewayRoundRobinDistributesRequests(t *testing.T) {
	t.Parallel()

	var countA atomic.Int64
	var countB atomic.Int64

	upstreamA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		countA.Add(1)
		w.Header().Set("X-Upstream-Id", "a")
		_, _ = io.WriteString(w, "a")
	}))
	t.Cleanup(upstreamA.Close)

	upstreamB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		countB.Add(1)
		w.Header().Set("X-Upstream-Id", "b")
		_, _ = io.WriteString(w, "b")
	}))
	t.Cleanup(upstreamB.Close)

	router, err := staticrouter.NewRouter([]string{upstreamA.URL, upstreamB.URL})
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

	for range 4 {
		resp, err := http.Get(srv.URL + "/api")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	}

	if countA.Load() != 2 || countB.Load() != 2 {
		t.Fatalf("counts a=%d b=%d, want 2 each", countA.Load(), countB.Load())
	}
}

func TestGatewaySingleUpstreamUnchanged(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter([]string{upstream.URL})
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

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if string(body) != "ok" {
		t.Fatalf("body = %q, want ok", string(body))
	}
}
