package reliability

import (
	"net/http"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	retry "github.com/mauriciomendonca/universal-api-gateway/internal/reliability/adapter/retry"
	timeout "github.com/mauriciomendonca/universal-api-gateway/internal/reliability/adapter/timeout"
)

// NewRoundTripper builds the upstream transport chain: timeout, then optional retry.
func NewRoundTripper(cfg config.ReliabilityConfig) http.RoundTripper {
	base := timeout.NewRoundTripper(http.DefaultTransport, cfg.UpstreamTimeout)
	if cfg.RetryMax <= 0 {
		return base
	}
	return retry.NewRoundTripper(base, cfg.RetryMax, cfg.RetryBackoff)
}
