package server

import (
	"encoding/json"
	"net/http"

	requestpipeline "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/adapter/request"
)

type authValidateBody struct {
	Subject string            `json:"subject"`
	Claims  map[string]string `json:"claims"`
}

func handleAuthValidate(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := toDomainRequest(r)

		resp, err := deps.Pipeline.Execute(r.Context(), req)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "pipeline error")
			return
		}
		if !shouldContinue(resp) {
			writeDomainResponse(w, resp)
			return
		}

		token := requestpipeline.BearerToken(req)
		identity, err := deps.Authenticator.Authenticate(r.Context(), token)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if identity.Claims == nil {
			identity.Claims = map[string]string{}
		}

		body, err := json.Marshal(authValidateBody{
			Subject: identity.Subject,
			Claims:  identity.Claims,
		})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "pipeline error")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}
