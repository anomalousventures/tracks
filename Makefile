# Makefile for Tracks Framework
# This provides convenient commands for development and CI

# Version information
VERSION := $(shell git --no-pager describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git --no-pager rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell git --no-pager log -1 --format=%cI 2>/dev/null || echo "unknown")
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: help lint lint-md lint-md-fix lint-go lint-js lint-js-fix format format-check install-linters

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

# JavaScript linting
lint-js: ## Run ESLint on JavaScript/TypeScript files
	@echo "Running ESLint..."
	@pnpm run --dir website lint:js

lint-js-fix: ## Auto-fix ESLint issues where possible
	@echo "Auto-fixing JavaScript issues..."
	@pnpm run --dir website lint:js:fix

# Formatting
format: ## Format code with Prettier
	@echo "Formatting code with Prettier..."
	@pnpm run --dir website format

format-check: ## Check code formatting with Prettier
	@echo "Checking code formatting..."
	@pnpm run --dir website format:check

# Aggregate linting target
lint: lint-md lint-go lint-js ## Run all linters

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
	@mkdir -p bin
	@go build $(LDFLAGS) -o bin/tracks ./cmd/tracks

build-mcp: ## Build tracks-mcp server
	@echo "Building tracks-mcp..."
	@mkdir -p bin
	@go build $(LDFLAGS) -o bin/tracks-mcp ./cmd/tracks-mcp

build-all: build build-mcp ## Build all binaries for current platform

# Cross-platform build targets
.PHONY: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows build-all-platforms

build-linux: ## Build for Linux amd64
	@echo "Building for Linux (amd64)..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/tracks-linux-amd64 ./cmd/tracks
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/tracks-mcp-linux-amd64 ./cmd/tracks-mcp

build-linux-arm64: ## Build for Linux arm64 (Raspberry Pi, WSL on ARM)
	@echo "Building for Linux (arm64)..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/tracks-linux-arm64 ./cmd/tracks
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/tracks-mcp-linux-arm64 ./cmd/tracks-mcp

build-darwin: ## Build for macOS amd64 (Intel)
	@echo "Building for macOS (amd64/Intel)..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/tracks-darwin-amd64 ./cmd/tracks
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/tracks-mcp-darwin-amd64 ./cmd/tracks-mcp

build-darwin-arm64: ## Build for macOS arm64 (Apple Silicon)
	@echo "Building for macOS (arm64/Apple Silicon)..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/tracks-darwin-arm64 ./cmd/tracks
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/tracks-mcp-darwin-arm64 ./cmd/tracks-mcp

build-windows: ## Build for Windows amd64
	@echo "Building for Windows (amd64)..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/tracks-windows-amd64.exe ./cmd/tracks
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/tracks-mcp-windows-amd64.exe ./cmd/tracks-mcp

build-all-platforms: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows ## Build for all platforms

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
.PHONY: changelog release-check release-prep release-tag release-rollback release-dry-run

changelog: ## Generate changelog with git-chglog
	@echo "Generating changelog..."
	@go tool git-chglog -o CHANGELOG.md
	@echo "✅ CHANGELOG.md generated"
	@echo "⚠️  Review the changelog and create a PR to commit it before tagging!"

release-check: ## Verify prerequisites for release
	@echo "Checking release prerequisites..."
	@echo "Checking git status..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ Working tree is dirty. Commit or stash changes first."; \
		exit 1; \
	fi
	@echo "✅ Working tree is clean"
	@echo "Checking branch..."
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then \
		echo "❌ Not on main branch. Switch to main first."; \
		exit 1; \
	fi
	@echo "✅ On main branch"
	@echo "Checking if CHANGELOG.md exists..."
	@if [ ! -f CHANGELOG.md ]; then \
		echo "❌ CHANGELOG.md not found. Run 'make changelog' first."; \
		exit 1; \
	fi
	@echo "✅ CHANGELOG.md exists"
	@echo "Running tests..."
	@$(MAKE) test
	@echo "✅ Tests passed"
	@echo "Running linters..."
	@$(MAKE) lint
	@echo "✅ Linters passed"
	@echo ""
	@echo "✅ All prerequisites passed! Ready to release."

release-prep: release-check ## Prepare for release (run checks and show next steps)
	@echo ""
	@echo "════════════════════════════════════════════════════════════════"
	@echo "  Release Preparation Complete"
	@echo "════════════════════════════════════════════════════════════════"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Ensure CHANGELOG.md has been merged to main"
	@echo "  2. Run: make release-tag VERSION=v0.x.0"
	@echo "  3. Monitor workflow: gh run watch <RUN_ID>"
	@echo "  4. Review draft release: gh release view v0.x.0"
	@echo "  5. Publish: gh release edit v0.x.0 --draft=false"
	@echo ""
	@echo "See .github/RELEASE_CHECKLIST.md for full checklist"
	@echo "════════════════════════════════════════════════════════════════"

release-tag: ## Create and push release tag (use VERSION=vX.Y.Z)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required."; \
		echo "Usage: make release-tag VERSION=v0.1.0"; \
		exit 1; \
	fi
	@echo "Creating release tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "✅ Tag $(VERSION) created locally"
	@echo "Pushing tag to origin..."
	@git push origin $(VERSION)
	@echo "✅ Tag $(VERSION) pushed to origin"
	@echo ""
	@echo "🚀 Release workflow triggered!"
	@echo "Monitor progress: gh run watch $$(gh run list --workflow=release.yml --limit 1 --json databaseId --jq '.[0].databaseId')"

release-rollback: ## Delete failed release and tag (use VERSION=vX.Y.Z)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required."; \
		echo "Usage: make release-rollback VERSION=v0.1.0"; \
		exit 1; \
	fi
	@echo "Rolling back release $(VERSION)..."
	@echo "Deleting draft release..."
	@gh release delete $(VERSION) --yes || echo "⚠️  Draft release not found or already deleted"
	@echo "Deleting local tag..."
	@git tag -d $(VERSION) || echo "⚠️  Local tag not found"
	@echo "Deleting remote tag..."
	@git push origin :refs/tags/$(VERSION) || echo "⚠️  Remote tag not found"
	@echo "✅ Rollback complete for $(VERSION)"
	@echo ""
	@echo "Fix the issues, then retry with: make release-tag VERSION=$(VERSION)"

release-dry-run: ## Test release process locally (doesn't publish)
	@echo "Running GoReleaser in dry-run mode..."
	@go tool goreleaser release --snapshot --clean --skip publish
	@echo "✅ Dry-run complete. Check dist/ directory for artifacts."

# Development targets
.PHONY: dev clean clean-platforms install deps

dev: ## Run tracks in development mode
	@echo "Starting development mode..."
	@go run $(LDFLAGS) ./cmd/tracks

install: build ## Install tracks CLI to $GOPATH/bin
	@echo "Installing tracks..."
	@go install ./cmd/tracks

deps: ## Download and tidy Go dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

clean-platforms: ## Clean cross-platform build artifacts only
	@echo "Cleaning cross-platform binaries..."
	@rm -f bin/tracks-linux-amd64 bin/tracks-mcp-linux-amd64
	@rm -f bin/tracks-linux-arm64 bin/tracks-mcp-linux-arm64
	@rm -f bin/tracks-darwin-amd64 bin/tracks-mcp-darwin-amd64
	@rm -f bin/tracks-darwin-arm64 bin/tracks-mcp-darwin-arm64
	@rm -f bin/tracks-windows-amd64.exe bin/tracks-mcp-windows-amd64.exe
	@echo "Cross-platform binaries cleaned!"

clean: ## Clean all build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -f coverage.out coverage.html
	@rm -rf website/build website/.docusaurus
	@echo "Clean complete!"
