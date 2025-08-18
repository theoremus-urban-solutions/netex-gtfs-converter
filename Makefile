# NETEX GTFS Conveter Go - Makefile
.PHONY: help build test lint fmt vet clean install dev-tools benchmark coverage security release-build release release-dry-run

# Variables
GO := go
BINARY_NAME := netex-gtfs-converter
BUILD_DIR := bin
COVERAGE_DIR := coverage
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

# Default target
help: ## Show this help message
	@echo "NETEX GTFS Conveter Go - Available commands:"
	@echo
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev-tools: ## Install development tools
	@echo "Installing development tools..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest

build: ## Build the CLI binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/netex-gtfs-converter
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/netex-gtfs-converter
	@GOOS=linux GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/netex-gtfs-converter
	@GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/netex-gtfs-converter
	@GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/netex-gtfs-converter
	@GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/netex-gtfs-converter
	@echo "Built all platform binaries in $(BUILD_DIR)/"

install: ## Install the CLI binary
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install -ldflags="$(LDFLAGS)" ./cmd/netex-gtfs-converter

# Testing
test: ## Run all tests
	@echo "Running tests..."
	@$(GO) test -race -v ./...

test-short: ## Run tests without race detection (faster)
	@echo "Running tests (short)..."
	@$(GO) test -v ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	@$(GO) test -race -v -count=1 ./...

benchmark: ## Run benchmark tests
	@echo "Running benchmarks..."
	@$(GO) test -bench=. -benchmem -count=3 ./...

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

# Code Quality
fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GO) vet ./...

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run --timeout=5m

lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running linter with auto-fix..."
	@golangci-lint run --fix --timeout=5m

staticcheck: ## Run staticcheck
	@echo "Running staticcheck..."
	@staticcheck ./...

security: ## Run security analysis
	@echo "Running security analysis..."
	@gosec -quiet -fmt=json -out=security-report.json ./...
	@gosec -quiet ./...

# Dependencies
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod verify

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@$(GO) get -u ./...
	@$(GO) mod tidy

deps-clean: ## Clean module cache
	@echo "Cleaning module cache..."
	@$(GO) clean -modcache

# Validation
validate: fmt vet lint test ## Run all validation checks

ci: validate benchmark ## Run CI pipeline locally

# Release
release-test: ## Test release build process
	@echo "Testing release build..."
	@$(MAKE) build-all
	@echo "Release test completed successfully"

release-prep: clean deps deps-update fmt lint staticcheck test build-all ## Prepare for release

release: ## Create a release with GoReleaser
	@echo "Creating release..."
	@goreleaser release --clean

release-dry-run: ## Dry run release with GoReleaser
	@echo "Dry running release..."
	@goreleaser release --clean --snapshot --skip=publish,announce

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f security-report.json
	@$(GO) clean -cache -testcache

clean-all: clean deps-clean ## Clean everything including caches

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@mkdir -p docs/api
	@$(GO) doc -all . > docs/api.txt
	@echo "Generating package documentation..."
	@$(GO) doc . > docs/package.txt
	@echo "Generating HTML documentation (if godoc is available)..."
	@which godoc >/dev/null 2>&1 && (godoc -http=:6060 & sleep 3 && curl -s http://localhost:6060/pkg/github.com/theoremus-urban-solutions/netex-gtfs-converter/ > docs/api/index.html && pkill -f "godoc -http=:6060") || echo "godoc not found - skipping HTML generation"
	@echo "Documentation generated:"
	@echo "  - All APIs: docs/api.txt"
	@echo "  - Package: docs/package.txt"
	@echo "  - HTML: docs/api/index.html (if godoc was available)"

docs-serve: ## Start documentation server
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Visit: http://localhost:6060/pkg/github.com/theoremus-urban-solutions/netex-gtfs-converter/"
	@godoc -http=:6060

docs-clean: ## Clean generated documentation
	@echo "Cleaning documentation..."
	@rm -rf docs/api/
	@rm -f docs/api.txt docs/package.txt

.FORCE: # Force target to always run

# Development workflow
dev: fmt lint test ## Quick development check

check: validate benchmark coverage security ## Comprehensive check before PR

# Docker (if needed in the future)
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t netex-gtfs-converter:$(VERSION) .

# Version info
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Go version: $(shell $(GO) version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

# Examples
example-basic: build ## Run basic example
	@echo "Running basic example..."
	@cd examples/basic && $(GO) run main.go

example-advanced: build ## Run advanced example  
	@echo "Running advanced example..."
	@cd examples/advanced && $(GO) run main.go

example-server: build ## Run API server example
	@echo "Running API server example..."
	@cd examples/api-server && $(GO) run main.go

# Statistics
stats: ## Show project statistics
	@echo "Project Statistics:"
	@echo "Lines of code: $(shell find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1 | awk '{print $$1}')"
	@echo "Go files: $(shell find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "Packages: $(shell $(GO) list ./... | wc -l)"
	@echo "Test files: $(shell find . -name '*_test.go' -not -path './vendor/*' | wc -l)"