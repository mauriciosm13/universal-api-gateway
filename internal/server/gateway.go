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

	route, err := h.deps.Router.Resolve(r.Context(), req)
	if errors.Is(err, routingport.ErrNoRoute) {
		http.Error(w, "no route matched", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "routing error", http.StatusInternalServerError)
		return
	}

	proxy, err := h.proxyFor(route.Upstream)
	if err != nil {
		http.Error(w, "invalid upstream", http.StatusInternalServerError)
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
