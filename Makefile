.PHONY: install build clean test help

# Build with CGO disabled (required for macOS compatibility)
GOFLAGS ?= -trimpath
LDFLAGS ?=

install: ## Install coauthor-cleaner binary
	CGO_ENABLED=0 go install $(GOFLAGS) -ldflags "$(LDFLAGS)" ./cmd/coauthor-cleaner

build: ## Build coauthor-cleaner binary to ./bin/
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o ./bin/coauthor-cleaner ./cmd/coauthor-cleaner

clean: ## Remove build artifacts
	rm -rf bin/
	go clean

test: ## Run tests
	go test -v ./...

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
