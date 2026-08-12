package server

import "net/http"

func newRoutes(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	gateway := newGatewayHandler(deps)

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /health/live", handleLive)
	mux.HandleFunc("GET /health/ready", handleReady)
	mux.Handle("/", gateway)

	return mux
}
