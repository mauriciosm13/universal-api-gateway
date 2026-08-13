package server

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	routingport "github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

type gatewayHandler struct {
	deps    Dependencies
	proxies sync.Map
}

func newGatewayHandler(deps Dependencies) *gatewayHandler {
	return &gatewayHandler{deps: deps}
}

func (h *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := toDomainRequest(r)

	resp, err := h.deps.Pipeline.Execute(r.Context(), req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "pipeline error")
		return
	}
	if !shouldContinue(resp) {
		writeDomainResponse(w, resp)
		return
	}

	route, err := h.deps.Router.Resolve(r.Context(), req)
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

	proxy.ServeHTTP(w, r)
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
	actual, _ := h.proxies.LoadOrStore(upstream, proxy)
	return actual.(*httputil.ReverseProxy), nil
}
