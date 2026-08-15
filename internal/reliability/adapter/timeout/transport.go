package timeout

import (
	"context"
	"net/http"
	"time"
)

// RoundTripper applies a per-request deadline to the wrapped transport.
type RoundTripper struct {
	Base    http.RoundTripper
	Timeout time.Duration
}

// NewRoundTripper returns a transport that enforces timeout on each request.
func NewRoundTripper(base http.RoundTripper, timeout time.Duration) *RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &RoundTripper{Base: base, Timeout: timeout}
}

// RoundTrip implements http.RoundTripper.
func (t *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), t.Timeout)
	defer cancel()

	type result struct {
		resp *http.Response
		err  error
	}

	ch := make(chan result, 1)
	go func() {
		resp, err := t.Base.RoundTrip(req.WithContext(ctx))
		ch <- result{resp: resp, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return res.resp, res.err
	}
}
