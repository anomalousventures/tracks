# Tracks

<p align="center">
  <img src="website/static/img/logo.png" alt="Tracks Logo" width="400">
</p>

<p align="center">
  <strong>⚡ Go, fast. A batteries included toolkit for hypermedia servers</strong>
</p>

<p align="center">
  <a href="https://anomalousventures.github.io/tracks/">Docs</a> ·
  <a href="#current-status">Status</a> ·
  <a href="#vision">Vision</a> ·
  <a href="#roadmap">Roadmap</a> ·
  <a href="#documentation">Documentation</a> ·
  <a href="#license">License</a>
</p>

---

## Current Status

Tracks is in **Phase 0 (Foundation)** development. The CLI tool and project scaffolding are being built.

**What works now:**

- ✅ Project structure and monorepo setup
- ✅ Documentation and roadmap
- ✅ CLI infrastructure (complete - v0.1.0 ready)
  - Root command with version, help, and global flags
  - Renderer pattern (Console, JSON, TUI-ready)
  - Mode detection (TTY, CI, flags, env vars)
  - Theme system with Lip Gloss styling
  - Comprehensive test coverage (unit + integration)
- ✅ Project generation (`tracks new` command)
  - Production-ready project scaffolding
  - Choice of database drivers (LibSQL, SQLite3, PostgreSQL)
  - Clean architecture with testable code
  - Auto-generated `.env` with sensible defaults
  - Docker Compose for local development
  - GitHub Actions CI workflow templates
  - `make dev` auto-starts required services
  - Cross-platform support (Linux, macOS, Windows)

## Quick Start

Install Tracks and create your first project:

```bash
# Install (requires Go 1.25+)
go install github.com/anomalousventures/tracks/cmd/tracks@latest

# Create a new project
tracks new myapp

# Start development (auto-starts Docker services and generates .env)
cd myapp
make dev    # Auto-starts required services, starts server with live reload

# Verify health endpoint
curl http://localhost:8080/api/health
```

**What you get:**

- Production-ready project structure with clean architecture
- Docker Compose for local development (auto-started with `make dev`)
- Auto-generated `.env` with sensible defaults
- GitHub Actions CI workflow ready to use
- All tests passing out of the box

See the [CLI documentation](https://anomalousventures.github.io/tracks/cli/new) and [Getting Started guide](https://anomalousventures.github.io/tracks/getting-started/installation) for detailed instructions.

**Coming next:**

- Phase 1: Code generation (resources, handlers, services)
- Phase 2: Authentication and authorization
- See [Roadmap](#roadmap) for details

## Vision

Tracks will be a code-generating web framework for Go that produces idiomatic, production-ready applications. Built for developers who want the productivity of modern frameworks with the performance and simplicity of Go.

### Design Principles

#### Type-Safe Everything

- [templ](https://github.com/a-h/templ) for compile-time HTML safety
- [SQLC](https://sqlc.dev/) generates type-safe Go from SQL
- No runtime reflection in generated code

#### Hypermedia-First

- HTMX integration for dynamic UIs without JavaScript
- Server-rendered templates with progressive enhancement
- RESTful URL patterns with hypermedia (HTML) responses, not JSON APIs

#### Idiomatic Go

- Generated code looks hand-written by an experienced Go developer
- Standard library patterns and clear error handling
- Dependency injection via interfaces for easy testing

#### Batteries Included

- Authentication (magic links, OTP, OAuth)
- Authorization (Casbin RBAC)
- Observability (OpenTelemetry)
- Development tooling (hot reload, linting, CI/CD templates)

## Roadmap

Tracks development is organized into 7 phases. See [`docs/roadmap`](./docs/roadmap) for detailed plans.

### Phase 0: Foundation (Current)

**Goal:** Working CLI that generates project scaffolds

**Epics:**

1. [CLI Infrastructure](./docs/roadmap/phases/0-foundation/epics/1-cli-infrastructure.md) - Cobra framework, Renderer pattern, version tracking
2. [Template Engine](./docs/roadmap/phases/0-foundation/epics/2-template-engine.md) - Embed system for project templates
3. [Project Generation](./docs/roadmap/phases/0-foundation/epics/3-project-generation.md) - `tracks new` command
4. [Generated Project Tooling](./docs/roadmap/phases/0-foundation/epics/4-generated-tooling.md) - Makefiles, Air, linting, Docker, CI/CD
5. [Documentation](./docs/roadmap/phases/0-foundation/epics/5-documentation.md) - Installation guides, getting started

**Status:** In progress · **Target:** v0.1.0

### Future Phases

- **Phase 1:** Core Web Layer - Chi router, handlers, middleware, templ templates
- **Phase 2:** Database Layer - SQLC, Goose migrations, LibSQL/PostgreSQL support
- **Phase 3:** Authentication - Magic links, OTP, OAuth providers
- **Phase 4:** Interactive TUI - Bubble Tea interface for generators
- **Phase 5:** Production - OpenTelemetry, health checks, deployment
- **Phase 6:** Authorization - Casbin RBAC
- **Phase 7:** Advanced Features - Real-time, file uploads, background jobs

[View Full Roadmap →](./docs/roadmap/README.md)

## Documentation

### Live Documentation

**[📚 View Full Documentation →](https://anomalousventures.github.io/tracks/)**

The complete documentation site includes guides, tutorials, API references, and examples.

### Product Requirements

Detailed PRD documents describe the complete vision:

- [Summary](./docs/prd/0_summary.md) - Overview and goals
- [Core Architecture](./docs/prd/1_core_architecture.md) - Application structure and patterns
- [Database Layer](./docs/prd/2_database_layer.md) - SQLC, migrations, and query patterns
- [Authentication](./docs/prd/3_authentication.md) - Magic links, OTP, and OAuth
- [Authorization (RBAC)](./docs/prd/4_authorization_rbac.md) - Casbin integration
- [Code Generation](./docs/prd/14_code_generation.md) - Generator design and CLI commands
- [Testing Strategy](./docs/prd/13_testing.md) - Unit, integration, and E2E testing
- [Deployment](./docs/prd/17_deployment.md) - Production deployment strategies

### Development

For contributors:

- [CONTRIBUTING.md](./CONTRIBUTING.md) - Development setup and coding standards
- [Roadmap](./docs/roadmap/README.md) - Phase breakdown and epic details
- [GitHub Issues](https://github.com/anomalousventures/tracks/issues) - Task tracking

#### Running Tests

Tracks uses a multi-layered testing approach for maximum cross-platform coverage:

```bash
make test                            # Unit tests (fast, all platforms)
go test ./tests/integration          # Integration tests (all platforms)
go test -tags=docker ./tests/integration  # Docker E2E tests (Ubuntu only)
make test-coverage                   # All tests with coverage reports
```

**Test Types:**

- **Unit Tests** - Colocated with source, run with `-short` flag, use race detector
- **Integration Tests** - File generation, validation, git operations (no external services)
- **Docker E2E Tests** - Full end-to-end with databases (Postgres, LibSQL) via Docker Compose

Docker E2E tests use the `//go:build docker` tag and only run on Ubuntu in CI due to Docker setup complexity on macOS/Windows. Non-Docker tests run on all platforms.

See [Testing Strategy PRD](./docs/prd/13_testing.md) for complete details.

## Philosophy

### Idiomatic Go

Tracks generates code that looks like it was hand-written by an experienced Go developer:

- No magic or reflection in generated code
- Standard library patterns and interfaces
- Explicit error handling with context
- Interface-based design for testability

### Type Safety

Type safety catches errors at compile time:

- templ templates are type-checked Go code
- SQLC generates Go structs and functions from SQL
- No string-based query building or SQL injection risks
- DTOs with field-level validation

### Hypermedia-First

HTMX enables rich interactions without JavaScript complexity:

- Server renders HTML with templ templates
- HTMX provides dynamic UI updates via HTML over the wire
- Progressive enhancement with Alpine.js where needed
- RESTful URL patterns with hypermedia (HTML) responses, not JSON APIs

### Convention Over Configuration

Sensible defaults with escape hatches:

- Standard project structure and file organization
- Generators follow consistent patterns
- Override defaults when needed
- `.tracks.yaml` for CLI metadata (committed), `.env` for runtime config (secrets)
- Hierarchical config: defaults → .env → environment variables

## Technology Stack

**Router:** Chi - lightweight, idiomatic HTTP router
**Templates:** templ - type-safe HTML templates compiled to Go
**Database:** SQLC - generates type-safe Go from SQL queries
**Migrations:** Goose - version-controlled database migrations
**Auth:** Custom + OAuth providers (Google, GitHub, etc.)
**Authorization:** Casbin - role-based access control
**Frontend:** HTMX + Alpine.js + TailwindCSS (optional)
**Observability:** OpenTelemetry - tracing, metrics, and logs
**Logging:** zerolog - structured JSON logging
**Testing:** Standard library + testify for assertions
**CLI/TUI:** Cobra + Bubble Tea (Charm stack)
**Development:** Docker Compose - local service orchestration

## Project Status

Tracks is in **pre-alpha development** (Phase 0). The API and generated code structure may change significantly before v1.0.

**Not ready for production use.** Follow development progress via [GitHub Issues](https://github.com/anomalousventures/tracks/issues) and [Roadmap](./docs/roadmap/README.md).

## Contributing

Contributions are welcome! The project is in active early development.

**Ways to contribute:**

- Report bugs and request features via [GitHub Issues](https://github.com/anomalousventures/tracks/issues)
- Discuss ideas in [GitHub Discussions](https://github.com/anomalousventures/tracks/discussions)
- Submit pull requests (see [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup)

## License

Tracks is released under the [MIT License](./LICENSE).

## Acknowledgments

Tracks builds on excellent open-source projects:

- [Chi](https://github.com/go-chi/chi) - HTTP routing
- [templ](https://github.com/a-h/templ) - Type-safe templates
- [SQLC](https://sqlc.dev/) - SQL code generation
- [Casbin](https://casbin.org/) - Authorization framework
- [Goose](https://github.com/pressly/goose) - Database migrations
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework

---

An [Anomalous Venture](https://github.com/enterprises/anomalousventures) by [Aaron Ross](https://github.com/ashmortar)
