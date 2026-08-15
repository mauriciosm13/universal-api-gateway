package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	authctx "github.com/mauriciomendonca/universal-api-gateway/internal/auth/context"
	"github.com/mauriciomendonca/universal-api-gateway/internal/reliability"
	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

const upstreamUserIDHeader = "X-User-Id"

type gatewayHandler struct {
	deps      Dependencies
	proxies   sync.Map
	transport http.RoundTripper
}

func newGatewayHandler(deps Dependencies) *gatewayHandler {
	return &gatewayHandler{
		deps:      deps,
		transport: reliability.NewRoundTripper(deps.Config.Reliability.WithDefaults()),
	}
}

func (h *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := toDomainRequest(r)

	ctx, resp, err := h.deps.Pipeline.Execute(r.Context(), req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "pipeline error")
		return
	}
	if !shouldContinue(resp) {
		writeDomainResponse(w, resp)
		return
	}

	route, err := h.deps.Router.Resolve(ctx, req)
	if errors.Is(err, routingport.ErrNoRoute) {
		writeJSONError(w, http.StatusNotFound, "no route matched")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "routing error")
		return
	}

	proxy, err := h.proxyFor(route.Upstream)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "invalid upstream")
		return
	}

	r = r.WithContext(ctx)
	applyUpstreamIdentity(r)
	proxy.ServeHTTP(w, r)
}

func applyUpstreamIdentity(r *http.Request) {
	if r.Header.Get(upstreamUserIDHeader) != "" {
		return
	}

	identity, ok := authctx.IdentityFrom(r.Context())
	if !ok || identity.Subject == "" {
		return
	}

	r.Header.Set(upstreamUserIDHeader, identity.Subject)
}

func (h *gatewayHandler) proxyFor(upstream string) (*httputil.ReverseProxy, error) {
	if cached, ok := h.proxies.Load(upstream); ok {
		return cached.(*httputil.ReverseProxy), nil
	}

	target, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = h.transport
	proxy.ErrorHandler = h.handleUpstreamError
	actual, _ := h.proxies.LoadOrStore(upstream, proxy)
	return actual.(*httputil.ReverseProxy), nil
}

func (h *gatewayHandler) handleUpstreamError(w http.ResponseWriter, _ *http.Request, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		writeJSONError(w, http.StatusGatewayTimeout, "upstream timeout")
	case isTimeoutError(err):
		writeJSONError(w, http.StatusGatewayTimeout, "upstream timeout")
	case isConnectionError(err):
		writeJSONError(w, http.StatusBadGateway, "upstream unavailable")
	default:
		writeJSONError(w, http.StatusBadGateway, "upstream unavailable")
	}
}

func isTimeoutError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func isConnectionError(err error) bool {
	var opErr *net.OpError
	return errors.As(err, &opErr)
}
