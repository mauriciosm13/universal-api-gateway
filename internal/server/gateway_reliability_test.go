package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
)

func TestGatewayUpstreamTimeoutReturns504(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
		Reliability: config.ReliabilityConfig{
			UpstreamTimeout: 50 * time.Millisecond,
			RetryMax:        0,
			RetryBackoff:    time.Millisecond,
		},
	}, router)

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/slow")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusGatewayTimeout, "upstream timeout")
}

func TestGatewayUpstreamRetrySucceeds(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, "recovered")
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
		Reliability: config.ReliabilityConfig{
			UpstreamTimeout: time.Second,
			RetryMax:        1,
			RetryBackoff:    time.Millisecond,
		},
	}, router)

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != "recovered" {
		t.Fatalf("body = %q, want recovered", string(body))
	}
	if calls.Load() != 2 {
		t.Fatalf("upstream calls = %d, want 2", calls.Load())
	}
}

func TestGatewayUpstreamRetrySkippedForPOST(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(upstream.Close)

	router, err := staticrouter.NewRouter(upstream.URL)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
		Reliability: config.ReliabilityConfig{
			UpstreamTimeout: time.Second,
			RetryMax:        2,
			RetryBackoff:    time.Millisecond,
		},
	}, router)

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api", http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	if calls.Load() != 1 {
		t.Fatalf("upstream calls = %d, want 1", calls.Load())
	}
}

func TestGatewayUpstreamConnectionFailureReturns502(t *testing.T) {
	t.Parallel()

	router, err := staticrouter.NewRouter("http://127.0.0.1:1")
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
		Reliability: config.ReliabilityConfig{
			UpstreamTimeout: 200 * time.Millisecond,
			RetryMax:        0,
			RetryBackoff:    time.Millisecond,
		},
	}, router)

	srv := httptest.NewServer(New(deps).httpServer.Handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	assertJSONError(t, resp, http.StatusBadGateway, "upstream unavailable")
}

func TestHandleUpstreamErrorHelpers(t *testing.T) {
	t.Parallel()

	h := &gatewayHandler{}

	rec := httptest.NewRecorder()
	h.handleUpstreamError(rec, nil, context.DeadlineExceeded)
	assertJSONErrorRecorder(t, rec, http.StatusGatewayTimeout, "upstream timeout")

	rec = httptest.NewRecorder()
	h.handleUpstreamError(rec, nil, &timeoutNetError{})
	assertJSONErrorRecorder(t, rec, http.StatusGatewayTimeout, "upstream timeout")

	rec = httptest.NewRecorder()
	var opErr net.OpError
	h.handleUpstreamError(rec, nil, &opErr)
	assertJSONErrorRecorder(t, rec, http.StatusBadGateway, "upstream unavailable")
}

func TestApplyUpstreamIdentityPreservesClientHeader(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "http://gateway/api", nil)
	req.Header.Set(upstreamUserIDHeader, "client-user")

	applyUpstreamIdentity(req)

	if got := req.Header.Get(upstreamUserIDHeader); got != "client-user" {
		t.Fatalf("header = %q, want client-user", got)
	}
}

type timeoutNetError struct{}

func (timeoutNetError) Error() string   { return "i/o timeout" }
func (timeoutNetError) Timeout() bool   { return true }
func (timeoutNetError) Temporary() bool { return true }

func assertJSONErrorRecorder(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
	}

	resp := rec.Result()
	defer resp.Body.Close()
	assertJSONError(t, resp, wantStatus, wantMessage)
}
