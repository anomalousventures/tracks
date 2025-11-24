# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## ⚠️ REQUIRED: Pre-Commit Validation

**Before making ANY commit, you MUST:**

1. **Run `make generate-mocks`** - Generate test mocks from interfaces (see [ADR-004](./docs/adr/004-mockery-for-test-mock-generation.md))
2. **Run `make lint`** - All linters must pass with zero errors
3. **Run `make test`** - Unit tests must pass with zero failures
4. **Run `make test-integration`** - Integration tests must pass with zero failures
5. **Remediate any errors** - Fix all issues found by linting and testing

**Failure to complete these steps successfully means the code is NOT ready to commit.**

**Note:** Generated mocks must be committed to the repository. The lint job checks that mocks are up-to-date.

**E2E and Docker testing:** Full E2E workflow and Docker container testing happens automatically in CI via GitHub Actions workflows. Use `make test-e2e-local` or `make test-docker-local` to test these workflows locally before pushing.

## ⚠️ CRITICAL: Command Line Tools

**This repository uses modern replacements for standard Unix tools. NEVER use the old commands:**

- **File search:** Use `fd` (NOT `find`)
- **Content search:** Use `rg` (NOT `grep`)

**Examples:**

```bash
# CORRECT:
fd "*.md"
rg "pattern" --type md

# WRONG - NEVER USE THESE:
find . -name "*.md"
grep -r "pattern" --include="*.md"
```

**Always use the Glob and Grep tools when possible instead of Bash commands for file/content search.**

## Quick Start

**For development setup, coding standards, and contribution guidelines, see [CONTRIBUTING.md](./CONTRIBUTING.md).**

This file contains Claude-specific context about the project architecture, common workflows, and patterns.

## Project Overview

Tracks is a code-generating web framework for Go that produces idiomatic, production-ready applications. It's a CLI tool built with Cobra that includes an interactive TUI (Bubble Tea) for code generation.

**Current Status:** Phase 1 (Core Web Layer) - Phase 0 (Foundation) complete. Epic 1.3 Phase 1 complete (basic asset infrastructure). Building asset pipeline, handler generation, middleware, routing helpers.

**Key Technologies:**

- CLI: Cobra + Bubble Tea (TUI)
- Generated Apps: Chi, templ, SQLC, HTMX
- Monorepo: Go + Docusaurus

## Quick Command Reference

```bash
make help              # Show all available commands
make test              # Run unit tests
make lint              # Run all linters
make build             # Build tracks CLI
make website-dev       # Start Docusaurus dev server
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for complete development workflow.

## Architecture

### Monorepo Structure

```text
tracks/
├── cmd/
│   ├── tracks/        # Main CLI tool
│   └── tracks-mcp/    # MCP server
├── internal/
│   ├── cli/           # CLI commands and UI
│   ├── generator/     # Code generators
│   └── templates/     # Embedded templates
├── docs/
│   ├── prd/           # Product requirements (detailed specs)
│   └── roadmap/       # Phase and epic breakdown
├── website/           # Docusaurus documentation
└── examples/          # Example generated apps
```

### Generated Application Architecture

Tracks generates applications with clean layered architecture:

**Request Flow:** HTTP Request → Handler → Service → Repository → Database

**Key Principles:**

1. **Dependency Injection** - Services receive dependencies via constructors for testability
2. **Interface-Based Design** - All external dependencies use interfaces
3. **Context Propagation** - Always pass `context.Context` as first parameter
4. **Explicit Error Handling** - Errors are wrapped with context using `fmt.Errorf("...: %w", err)`
5. **Type-Safe SQL** - Uses SQLC to generate Go code from SQL queries
6. **Type-Safe Templates** - Uses templ for compile-time HTML safety

**Layer Responsibilities:**

- **HTTP Layer** (`internal/http/`) - All web-facing code
  - **Server** (`server.go`) - HTTP server setup and dependency injection
  - **Routes** (`routes.go`) - Route registration and middleware chain
  - **Handlers** (`handlers/`) - HTTP request/response, validation using DTOs, orchestrate domain services (can use multiple domains via interfaces)
  - **Middleware** (`middleware/`) - Single-responsibility composable functions (auth, security, logging, etc.)
  - **Views** (`views/`) - templ components compiled to Go
- **Domain Services** (`internal/domain/*/service.go`) - Business logic, implements `interfaces.*Service`
- **Domain Repositories** (`internal/domain/*/repository.go`) - Data access, wraps `internal/db/generated` SQLC code, implements `interfaces.*Repository`
- **Domain DTOs** (`internal/domain/*/dto.go`) - Request/response data transfer objects
- **Interfaces** (`internal/interfaces/`) - Service and repository contracts (zero implementations, prevents import cycles)

### CLI Tool Architecture (tracks itself)

The tracks CLI tool follows the same clean architecture principles as generated applications, with some CLI-specific patterns.

**Core Principles:**

1. **Dependency Injection** - Commands receive dependencies via constructors
2. **Interface-First** - Interfaces defined in consumer packages (`internal/cli/interfaces/`)
3. **Context Propagation** - Logger and request-scoped values passed via context
4. **Separation of Concerns** - Clear boundaries between commands, validation, generation

**Dual-Output Strategy:**

- **Renderer** (stdout) - User-facing output using Lip Gloss/Bubbles (human-friendly)
- **Logger** (stderr) - Developer/debug logging using zerolog (controlled by `TRACKS_LOG_LEVEL`)

This separation keeps user experience clean while enabling debugging.

**Package Structure:**

```text
internal/cli/
├── commands/          # Command implementations (NewCommand, VersionCommand, etc.)
├── interfaces/        # Interfaces consumed by CLI (Validator, ProjectGenerator)
├── renderer/          # Output formatting (Console, JSON, TUI)
├── ui/                # Mode detection, theming
├── logger.go          # zerolog setup for developer logging
└── root.go            # Root command setup, DI wiring
```

**Key Files:**

- `internal/cli/commands/*.go` - Individual command implementations
- `internal/cli/interfaces/*.go` - Interfaces for external dependencies
- `internal/cli/ui/mode.go` - Mode detection (TTY, CI, flags)
- `internal/cli/ui/theme.go` - Lip Gloss styles
- `internal/cli/renderer/` - Renderer implementations (Console, JSON, TUI)
- `internal/cli/logger.go` - zerolog configuration

## Code Generation Principles

When implementing or reviewing generators:

1. Generated code should look hand-written by an experienced Go developer
2. No magic or reflection - everything is explicit
3. DTOs are generated with field-level validation rules
4. Services use dependency injection for easy testing
5. Route helpers are auto-generated for type-safe URLs
6. Tests with mocks are generated by default
7. **Template sequencing for mock-dependent tests** - See [ADR-008](./docs/adr/008-generator-template-sequencing.md)
   - Render app templates → `go mod tidy` → `go mod download all` → `make generate` → render test templates
   - This ensures test files can import generated mocks without circular dependency issues
   - Apply this pattern to ALL generators that create testable code (handlers, resources, repositories)

## Important Patterns and Gotchas

### Go 1.25+ Tool Directive

All development tools (golangci-lint, air, etc.) use the `go tool <name>` pattern:

```bash
go tool golangci-lint run
go tool air -c .air.toml
```

This is the modern Go 1.25+ pattern where tools are declared in `go.mod` with the `tool` directive and invoked via `go tool <name>`. Never suggest global installations.

### CLI Output vs Generated App Logging

- **CLI Tool (tracks)** - Uses Renderer pattern for human-friendly output (Lip Gloss, Bubbles)
- **Generated Apps** - Use zerolog for structured JSON logging in production

These are two different contexts with different needs. Don't confuse them.

### Cross-Platform Path Handling

Always use `filepath` package for path operations:

```go
// GOOD: Cross-platform
projectDir := filepath.Join(baseDir, projectName)
templatePath := filepath.FromSlash("internal/templates/project")

// BAD: Platform-specific
projectDir := baseDir + "/" + projectName
```

### Error Wrapping

Always wrap errors with context using `%w`:

```go
// GOOD: Preserves error chain
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// BAD: Loses error chain
if err != nil {
    return errors.New("error occurred")
}
```

## Database Context

- **Default:** LibSQL/Turso (requires CGO, gcc/musl-dev on Alpine)
- **Alternatives:** SQLite (requires CGO), PostgreSQL (no CGO, static builds)
- **Migrations:** Goose with timestamp prefixes
- **Queries:** Written in SQL, processed by SQLC for type safety
- **IDs:** UUIDv7 (timestamp-ordered UUIDs)

## Documentation Structure

- **`/docs/prd/`** - Detailed product requirements (the "what" and "why")
- **`/docs/roadmap/`** - Phase breakdown and epic planning (the "when" and "how")
- **`/website/docs/`** - User-facing documentation (guides, tutorials)
- **`CONTRIBUTING.md`** - Development setup and standards (start here for dev work)

## Configuration

Tracks uses separate configuration files for different purposes. See [ADR-007](./docs/adr/007-configuration-file-separation.md) for complete details.

### CLI Project Metadata (`.tracks.yaml`)

**Purpose:** Machine-readable metadata for Tracks CLI commands
**Committed to Git:** ✅ Yes
**Contains secrets:** ❌ No

Contains: Database driver, module path, project name, version info, resource registry (future)

Read by: `tracks` CLI commands (`new`, `generate`, `db migrate`, `upgrade`)

### Application Runtime Configuration (`.env`)

**Purpose:** Environment-specific runtime configuration for generated applications
**Committed to Git:** ❌ No (`.env.example` is committed as template)
**Contains secrets:** ✅ Yes (database URLs, session keys, API credentials)

Hierarchical configuration (lowest to highest priority):

1. Default values in code (`viper.SetDefault()`)
2. `.env` file (development only, gitignored)
3. Environment variables (production, prefixed with `APP_`)

Generated applications do NOT read `.tracks.yaml` for runtime configuration.

## Development Workflow Tips

### Working on CLI Features

1. Read the relevant epic in `/docs/roadmap/phases/` to understand the plan
2. Check GitHub issues for task breakdown and acceptance criteria
3. Run `make build && ./bin/tracks <command>` to test changes
4. Use `make lint` before committing

### Working on Generators

1. Check `/docs/prd/` for detailed specs on what should be generated
2. Templates live in `internal/templates/` and are embedded via `embed.go`
3. Test by generating actual projects and verifying they build/run
4. Generated code should pass `go vet` and `golangci-lint`

### Working on Documentation

- Markdown files must pass `make lint-md`
- Roadmap docs live in `/docs/roadmap/`
- PRD docs live in `/docs/prd/`
- User docs live in `/website/docs/` (Docusaurus)

### Architecture Tests

We use architecture tests to enforce design principles programmatically. These tests run on every `make test` and prevent architectural drift.

**Current Architecture Tests:**

1. **Import Cycle Detection** (`TestNoImportCycles`)
   - **Purpose:** Prevent circular dependencies
   - **Method:** Uses `go list -json` to detect cycles
   - **Location:** `internal/cli/commands/architecture_test.go`
   - **References:** [ADR-002](./docs/adr/002-interface-placement-consumer-packages.md)

2. **Interface Location Validation** (`TestInterfacesPackageOnlyContainsInterfaces`)
   - **Purpose:** Enforce "interfaces in consumer packages" rule
   - **Method:** AST parsing to verify only interfaces in `cli/interfaces/`
   - **Location:** `internal/cli/interfaces/interfaces_test.go`
   - **References:** [ADR-002](./docs/adr/002-interface-placement-consumer-packages.md)

3. **DI Pattern Enforcement** (`TestCommandsUseDI`)
   - **Purpose:** Ensure all commands use dependency injection
   - **Method:** Pattern matching to verify each `*Command` has `New*Command` constructor
   - **Location:** `internal/cli/commands/architecture_test.go`
   - **References:** [ADR-001](./docs/adr/001-dependency-injection-for-cli-commands.md)

**When to Add Architecture Tests:**

- New architectural rule that can be verified programmatically
- Pattern that must be followed by all new code
- Design decision that prevents future bugs if enforced
- When code review alone isn't sufficient to catch violations

**How to Add a New Architecture Test:**

1. Add test function to appropriate `architecture_test.go` file
2. Use `go list`, AST parsing, or pattern matching as appropriate
3. Provide clear error messages that explain the violation and reference relevant ADRs
4. Verify test passes on current codebase
5. Document the test in this section of CLAUDE.md

## Testing Strategy

Tracks uses a multi-layered testing approach to ensure code quality while maximizing cross-platform coverage.

### Test Organization

**Unit Tests** (`*_test.go` colocated with source):

- Fast, isolated tests with no external dependencies
- Run with `-short` flag
- Use the race detector (`-race`)
- Execute on ALL platforms (Ubuntu, macOS, Windows)

**Integration Tests** (`tests/integration/`):

- Test component integration without external services
- Include file generation, validation, git operations
- Execute on ALL platforms (Ubuntu, macOS, Windows)

**Docker E2E Tests** (`tests/integration/` with `//go:build docker` tag):

- Full end-to-end tests requiring Docker Compose
- Test generated projects with databases (Postgres, LibSQL)
- Execute ONLY on Ubuntu runners (Docker setup issues on macOS/Windows)

### Build Tags for Test Separation

Use Go build tags to separate Docker-requiring tests:

```go
//go:build docker

package integration

import "testing"

func TestE2E_Postgres(t *testing.T) {
    runE2ETest(t, "postgres")
}
```

**Running tests locally:**

```bash
make test                           # Unit tests only (short)
make test-integration               # Integration tests (Go tests in tests/integration/)
make test-coverage                  # Unit + integration with coverage reports
make test-e2e-local                 # Test E2E workflow locally (mimics CI)
make test-docker-local              # Test Docker workflow locally (mimics CI)
```

**Local E2E Testing:**

- **`make test-e2e-local`** - Tests developer workflow (fast, no Docker required)
  - **Use when:** Verifying generated projects build/test/run correctly
  - **Tests:** `tracks new` → `make test` → `make dev` → health check
  - **Database:** SQLite3 only (no external services needed)
  - **Speed:** ~30 seconds

- **`make test-docker-local`** - Tests containerization (slower, requires Docker)
  - **Use when:** Verifying Dockerfile, security scans, production deployment
  - **Tests:** `docker build` → Trivy scan → `docker run` → health check
  - **Database:** SQLite3 only (no docker-compose needed)
  - **Speed:** ~2-3 minutes (includes image build and scan)

Both targets mimic CI workflows but only test sqlite3 for speed. CI tests all three drivers (sqlite3, postgres, go-libsql).

**CI Test Jobs:**

1. **unit-tests** - Runs on all platforms with `-race -short`
2. **integration-tests** - Runs on all platforms (Go integration tests)
3. **e2e-workflow** - Tests developer workflow (`tracks new`, `make test`, `make dev`) for all database drivers
4. **docker-workflow** - Tests Docker containerization (`docker build`, Trivy scan, `docker run`) for all database drivers

Each job uploads separate coverage reports to Codecov with platform-specific flags.

### Timeout Configuration

Integration and E2E tests support platform-specific timeouts via environment variables:

- `INTEGRATION_TEST_SHORT_TIMEOUT` (default: 2s, macOS: 4s, Windows: 6s)
- `INTEGRATION_TEST_MEDIUM_TIMEOUT` (default: 10s, macOS: 20s, Windows: 30s)
- `INTEGRATION_TEST_LONG_TIMEOUT` (default: 15s, macOS: 30s, Windows: 45s)
- `INTEGRATION_TEST_E2E_TIMEOUT` (default: 120s, macOS: 180s, Windows: 240s)

CI sets these automatically based on runner OS.

## Commit and PR Guidelines

**See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed PR process.**

Quick reminders:

- PR titles use Conventional Commits format (becomes squash merge commit message)
- Follow the PR template in `.github/pull_request_template.md`
- Keep commit messages TERSE and focused
- Write PR descriptions naturally, avoid AI slop
- Use `make lint` before committing

## Environment Requirements

- **Go:** 1.25+ required
- **Node.js:** 24+ with pnpm 10+ (for website)
- **CGO:** Required for LibSQL/SQLite (needs gcc on Linux, xcode on macOS)
- **Git:** For version control and release process

## Release Process

```bash
make changelog              # Generate changelog from commits
make release-prep           # Verify prerequisites and show next steps
make release-tag VERSION=v0.1.0 # Create and push release tag
```

See [docs/RELEASE_PROCESS.md](./docs/RELEASE_PROCESS.md) for complete release workflow.

---

## 🚨 MANDATORY: Always Validate Before Committing

**This is a critical requirement that cannot be skipped:**

Before creating any commit, you must SUCCESSFULLY complete:

1. **`make generate-mocks`** - Generate test mocks
2. **`make lint`** - Fix ALL linting errors
3. **`make test`** - Fix ALL test failures (unit tests)
4. **`make test-integration`** - Fix ALL integration test failures

If any command fails, you MUST remediate the errors before proceeding. Code that fails linting or testing is NOT ready to commit.

E2E and Docker testing happens automatically in CI workflows.

---

**Note:** This file focuses on Claude-specific context. For development setup, coding standards, and contribution guidelines, see [CONTRIBUTING.md](./CONTRIBUTING.md).
