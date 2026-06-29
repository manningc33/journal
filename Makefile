BINARY := journal
PKG := github.com/manningc33/journal
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X $(PKG)/internal/cli.Version=$(VERSION)"

.PHONY: all build install test test-race cover fmt vet lint tidy clean

all: fmt vet test build

build: ## Build the journal binary into ./bin
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/journal

install: ## Install the journal binary to GOBIN
	go install $(LDFLAGS) ./cmd/journal

test: ## Run unit tests
	go test ./...

test-race: ## Run tests with the race detector
	go test -race ./...

cover: ## Run tests and open the coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt: ## Format all Go source
	gofmt -w .

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint if installed
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"

tidy: ## Tidy module dependencies
	go mod tidy

clean: ## Remove build artifacts
	rm -rf bin coverage.out
