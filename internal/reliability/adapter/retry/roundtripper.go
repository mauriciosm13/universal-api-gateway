package retry

import (
	"io"
	"net/http"
	"time"
)

// RoundTripper retries idempotent requests on transient failures.
type RoundTripper struct {
	Base     http.RoundTripper
	MaxRetry int
	Backoff  time.Duration
}

// NewRoundTripper wraps base with retry behavior. MaxRetry is the number of
// additional attempts after the first try.
func NewRoundTripper(base http.RoundTripper, maxRetry int, backoff time.Duration) *RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &RoundTripper{Base: base, MaxRetry: maxRetry, Backoff: backoff}
}

// RoundTrip implements http.RoundTripper.
func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	attempts := rt.MaxRetry + 1
	var lastErr error

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if !isIdempotent(req.Method) {
				break
			}
			time.Sleep(rt.Backoff * time.Duration(attempt))
			cloned := req.Clone(req.Context())
			if cloned.GetBody != nil && req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, err
				}
				cloned.Body = body
			}
			req = cloned
		}

		resp, err := rt.Base.RoundTrip(req)
		if err != nil {
			lastErr = err
			if attempt < attempts-1 && isIdempotent(req.Method) && shouldRetryError(err) {
				continue
			}
			return nil, err
		}

		if attempt < attempts-1 && isIdempotent(req.Method) && shouldRetryStatus(resp.StatusCode) {
			discardBody(resp.Body)
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func shouldRetryError(err error) bool {
	return err != nil
}

func shouldRetryStatus(code int) bool {
	switch code {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func discardBody(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
