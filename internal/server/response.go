package server

import (
	"net/http"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
)

func writeDomainResponse(w http.ResponseWriter, resp domain.Response) {
	for key, values := range resp.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if len(resp.Body) > 0 {
		_, _ = w.Write(resp.Body)
	}
}

func shouldContinue(resp domain.Response) bool {
	return resp.StatusCode == requestpipeline.ContinueStatus
}
