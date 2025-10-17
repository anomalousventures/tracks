# Makefile for Tracks Framework
# This provides convenient commands for development and CI

.PHONY: help lint lint-md lint-md-fix lint-go install-linters

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Install linters
install-linters: ## Install required linters (markdownlint-cli2)
	@echo "Installing markdownlint-cli2..."
	@which npm > /dev/null || (echo "Error: npm not found. Please install Node.js first." && exit 1)
	@npm install -g markdownlint-cli2
	@echo "Linters installed successfully!"

# Markdown linting
lint-md: ## Run markdown linter on all .md files
	@echo "Running markdown linter..."
	@pnpm run lint:md

lint-md-fix: ## Auto-fix markdown linting issues where possible
	@echo "Auto-fixing markdown issues..."
	@pnpm run lint:md:fix

# Go linting
lint-go: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@go tool golangci-lint run ./...

# Aggregate linting target
lint: lint-md lint-go ## Run all linters

# Go-related targets
.PHONY: test test-coverage test-integration test-all build build-all

test: ## Run Go unit tests
	@echo "Running unit tests..."
	@go test -v -short ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./...

test-all: test test-integration ## Run all tests

build: ## Build tracks CLI
	@echo "Building tracks..."
	@go build -o bin/tracks ./cmd/tracks

build-mcp: ## Build tracks-mcp server
	@echo "Building tracks-mcp..."
	@go build -o bin/tracks-mcp ./cmd/tracks-mcp

build-all: build build-mcp ## Build all binaries

# Website targets
.PHONY: website-dev website-build website-serve website-deploy

website-dev: ## Start Docusaurus dev server
	@echo "Starting website dev server..."
	@pnpm run website:dev

website-build: ## Build website for production
	@echo "Building website..."
	@pnpm run website:build

website-serve: ## Serve built website locally
	@echo "Serving website..."
	@pnpm run website:serve

website-deploy: website-build ## Deploy website to production
	@echo "Deploying website..."
	@pnpm run website:deploy

# Release targets
.PHONY: changelog release-dry-run release

changelog: ## Generate changelog with git-chglog
	@echo "Generating changelog..."
	@go tool git-chglog -o CHANGELOG.md

release-dry-run: ## Test release process locally
	@echo "Running GoReleaser in dry-run mode..."
	@go tool goreleaser release --snapshot --clean --skip publish

release: ## Create a new release (use VERSION=vX.Y.Z)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release VERSION=v0.1.0"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@go tool git-chglog -o CHANGELOG.md
	@git add CHANGELOG.md
	@git commit -m "chore: update changelog for $(VERSION)" || true
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "Release $(VERSION) created and pushed!"

# Development targets
.PHONY: dev clean install deps

dev: ## Run tracks in development mode
	@echo "Starting development mode..."
	@go run ./cmd/tracks

install: build ## Install tracks CLI to $GOPATH/bin
	@echo "Installing tracks..."
	@go install ./cmd/tracks

deps: ## Download and tidy Go dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -f coverage.out coverage.html
	@rm -rf website/build website/.docusaurus
	@echo "Clean complete!"
