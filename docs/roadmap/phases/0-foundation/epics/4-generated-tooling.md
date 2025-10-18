# Epic 4: Generated Project Tooling

[← Back to Phase 0](../0-foundation.md) | [← Epic 3](./3-project-generation.md) | [Epic 5 →](./5-documentation.md)

## Overview

Enhance generated projects with development and build tooling. This includes Makefiles, Air configuration for hot-reload, linting setup, and CI/CD templates. The goal is to make generated projects immediately productive for development and ready for production deployment.

## Goals

- Complete Makefile with standard targets for generated projects
- Air configuration for hot-reload during development
- Linting configuration (golangci-lint)
- .gitignore with sensible defaults
- Docker configuration templates
- CI/CD workflow templates

## Scope

### In Scope

- Makefile generation with targets: build, test, lint, dev, clean
- Air configuration (.air.toml) for hot-reload
- golangci-lint configuration
- .gitignore for Go projects
- Basic Dockerfile for containerization
- GitHub Actions CI workflow
- Helper scripts (build.sh, etc.)

### Out of Scope

- Kubernetes manifests - defer to Phase 5 (Production)
- Advanced CI/CD (release automation) - Phase 5
- Multi-stage build optimization - Phase 5
- Custom linting rules beyond defaults
- Pre-commit hooks - future enhancement

## Task Breakdown

The following tasks will become GitHub issues:

1. **Create Makefile template with standard targets**
2. **Add build target with proper Go flags**
3. **Add test target with coverage reporting**
4. **Add lint target using golangci-lint**
5. **Add dev target that uses Air**
6. **Add clean target for build artifacts**
7. **Add help target with target documentation**
8. **Create .air.toml template for hot-reload**
9. **Create .golangci.yml with sensible linting rules**
10. **Create comprehensive .gitignore for Go projects**
11. **Create basic Dockerfile template**
12. **Create GitHub Actions CI workflow template**
13. **Write integration tests for Makefile targets**
14. **Verify generated tooling works on all platforms**
15. **Document tooling setup and usage**

## Dependencies

### Prerequisites

- Epic 3 (Project Generation) - need projects to add tooling to
- Understanding of Makefile syntax
- Air, golangci-lint knowledge

### Blocks

- Developer productivity (Makefile, Air enable rapid development)
- CI/CD pipeline (GitHub Actions workflow enables automation)

## Acceptance Criteria

- [ ] Generated Makefile has all standard targets
- [ ] `make build` produces working binary
- [ ] `make test` runs tests and shows coverage
- [ ] `make lint` runs golangci-lint successfully
- [ ] `make dev` starts Air hot-reload server
- [ ] `make clean` removes build artifacts
- [ ] `make help` shows all targets with descriptions
- [ ] .air.toml enables hot-reload with correct paths
- [ ] .golangci.yml catches common Go issues
- [ ] .gitignore excludes build artifacts, IDE files, etc.
- [ ] Dockerfile builds minimal working container
- [ ] GitHub Actions workflow runs tests and linting
- [ ] All tooling works on Linux, macOS, and Windows (where applicable)
- [ ] Documentation explains each tool's purpose and usage

## Technical Notes

### Makefile Structure

```makefile
.PHONY: help build test lint dev clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/server cmd/server/main.go

test: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...

lint: ## Run linters
	go tool golangci-lint run

dev: ## Start development server with hot-reload
	air -c .air.toml

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out
```

### Air Configuration

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main cmd/server/main.go"
  bin = "tmp/main"
  include_ext = ["go", "templ"]
  exclude_dir = ["tmp", "vendor", "node_modules"]
  delay = 1000
```

### golangci-lint Configuration

Start with recommended linters, not too strict:

```yaml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
```

**Note:** The Makefile uses `go tool golangci-lint run` instead of calling `golangci-lint` directly. This follows Go 1.25's modern tooling pattern where tools are declared in `go.mod` with the `tool` directive and invoked via `go tool <name>`.

### Dockerfile

Keep it simple for Phase 0:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
```

### GitHub Actions

Basic CI workflow:

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: make test
      - run: make lint
```

## Testing Strategy

- Integration tests that run each Makefile target
- Verify Air configuration with actual hot-reload
- Test linting catches common issues
- Verify Docker build produces working container
- Test CI workflow in actual GitHub Actions

## Next Epic

[Epic 5: Documentation & Installation →](./5-documentation.md)
