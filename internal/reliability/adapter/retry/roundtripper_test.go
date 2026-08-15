package retry_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/reliability/adapter/retry"
)

type stubTransport struct {
	responses []*http.Response
	errs      []error
	calls     atomic.Int32
}

func (s *stubTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	idx := int(s.calls.Add(1)) - 1
	if idx < len(s.errs) && s.errs[idx] != nil {
		return nil, s.errs[idx]
	}
	if idx < len(s.responses) {
		return s.responses[idx], nil
	}
	return nil, errors.New("unexpected call")
}

func TestRoundTripperRetries503ThenSucceeds(t *testing.T) {
	t.Parallel()

	base := &stubTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(http.NoBody)},
			{StatusCode: http.StatusOK, Body: io.NopCloser(http.NoBody)},
		},
	}

	rt := retry.NewRoundTripper(base, 1, time.Millisecond)
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if base.calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", base.calls.Load())
	}
}

func TestRoundTripperDoesNotRetryPOST(t *testing.T) {
	t.Parallel()

	base := &stubTransport{
		errs: []error{errors.New("connection refused")},
	}

	rt := retry.NewRoundTripper(base, 2, time.Millisecond)
	req, err := http.NewRequest(http.MethodPost, "http://example.com", http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	_, err = rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error")
	}
	if base.calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", base.calls.Load())
	}
}

func TestRoundTripperRetriesNetworkError(t *testing.T) {
	t.Parallel()

	base := &stubTransport{
		errs: []error{
			errors.New("connection reset"),
			nil,
		},
		responses: []*http.Response{
			nil,
			{StatusCode: http.StatusOK, Body: io.NopCloser(http.NoBody)},
		},
	}

	rt := retry.NewRoundTripper(base, 1, time.Millisecond)
	req, err := http.NewRequest(http.MethodHead, "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	resp.Body.Close()

	if base.calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", base.calls.Load())
	}
}

func TestRoundTripperRetries502504(t *testing.T) {
	t.Parallel()

	for _, code := range []int{http.StatusBadGateway, http.StatusGatewayTimeout} {
		code := code
		t.Run(http.StatusText(code), func(t *testing.T) {
			t.Parallel()

			base := &stubTransport{
				responses: []*http.Response{
					{StatusCode: code, Body: io.NopCloser(http.NoBody)},
					{StatusCode: http.StatusOK, Body: io.NopCloser(http.NoBody)},
				},
			}

			rt := retry.NewRoundTripper(base, 1, time.Millisecond)
			req, err := http.NewRequest(http.MethodOptions, "http://example.com", nil)
			if err != nil {
				t.Fatalf("NewRequest() error = %v", err)
			}

			resp, err := rt.RoundTrip(req)
			if err != nil {
				t.Fatalf("RoundTrip() error = %v", err)
			}
			resp.Body.Close()

			if base.calls.Load() != 2 {
				t.Fatalf("calls = %d, want 2", base.calls.Load())
			}
		})
	}
}

func TestRoundTripperZeroRetries(t *testing.T) {
	t.Parallel()

	base := &stubTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusBadGateway, Body: io.NopCloser(http.NoBody)},
		},
	}

	rt := retry.NewRoundTripper(base, 0, time.Millisecond)
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", resp.StatusCode)
	}
	if base.calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", base.calls.Load())
	}
}

func TestRoundTripperExhaustsRetries(t *testing.T) {
	t.Parallel()

	base := &stubTransport{
		errs: []error{errors.New("fail"), errors.New("fail"), errors.New("fail")},
	}

	rt := retry.NewRoundTripper(base, 2, time.Millisecond)
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	if _, err := rt.RoundTrip(req); err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if base.calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", base.calls.Load())
	}
}

func TestRoundTripperRetriesWithGetBody(t *testing.T) {
	t.Parallel()

	body := strings.NewReader("payload")
	base := &stubTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(http.NoBody)},
			{StatusCode: http.StatusOK, Body: io.NopCloser(http.NoBody)},
		},
	}

	rt := retry.NewRoundTripper(base, 1, time.Millisecond)
	req, err := http.NewRequest(http.MethodGet, "http://example.com", body)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("payload")), nil
	}

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	resp.Body.Close()

	if base.calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", base.calls.Load())
	}
}
