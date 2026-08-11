package server

import "net/http"

type rootHandler struct {
	gateway *gatewayHandler
}

func newRootHandler(deps Dependencies) *rootHandler {
	return &rootHandler{
		gateway: newGatewayHandler(deps),
	}
}

func (h *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		switch r.URL.Path {
		case "/health":
			handleHealth(w, r)
			return
		case "/health/live":
			handleLive(w, r)
			return
		case "/health/ready":
			handleReady(w, r)
			return
		}
	}

	h.gateway.ServeHTTP(w, r)
}
