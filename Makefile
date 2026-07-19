# HazinaDB developer tasks. Run `make help` to list them.

.DEFAULT_GOAL := help
BINARY := hazinadb

.PHONY: run
run: ## Run the server (make run ARGS="--addr :9000")
	go run ./cmd/hazinadb $(ARGS)

.PHONY: build
build: ## Compile the binary into ./bin
	go build -o bin/$(BINARY) ./cmd/hazinadb

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: test-race
test-race: ## Run all tests with the race detector (needs a C compiler / CGO_ENABLED=1)
	CGO_ENABLED=1 go test -race ./...

.PHONY: fmt
fmt: ## Format all Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet static checks
	go vet ./...

.PHONY: check
check: fmt vet test ## Format, vet, and test — run before committing

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'
