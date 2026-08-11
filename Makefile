.PHONY: help run build test vet fmt fmt-check quality clean docker-build k8s-validate

GOBIN ?= $(shell go env GOPATH)/bin
GATEWAY_PORT ?= 8080

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Run the gateway locally
	GATEWAY_PORT=$(GATEWAY_PORT) go run ./cmd/gateway

build: ## Build gateway binary to bin/gateway
	go build -o bin/gateway ./cmd/gateway

test: ## Run all tests
	go test -race -count=1 ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format Go source files
	gofmt -w .

fmt-check: ## Fail if Go files are not formatted
	@test -z "$$(gofmt -l .)"

quality: ## Run the full engineering quality gate
	bash quality/scripts/quality-gate.sh

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out quality/reports

docker-build: ## Build Docker image
	docker build -t universal-api-gateway:local .

k8s-validate: ## Validate Kubernetes manifests with kustomize
	kubectl kustomize deploy/kubernetes/base >/dev/null
	kubectl kustomize deploy/kubernetes/overlays/dev >/dev/null
	kubectl kustomize deploy/kubernetes/overlays/staging >/dev/null
	kubectl kustomize deploy/kubernetes/overlays/prod >/dev/null

.DEFAULT_GOAL := help
