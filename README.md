# Tracks

<p align="center">
  <picture>
    <source type="image/webp"
            srcset="https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/tracks-logo-256.webp 256w,
                    https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/tracks-logo-512.webp 512w"
            sizes="(max-width: 768px) 256px, 400px">
    <img src="https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/tracks-logo-512.png"
         srcset="https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/tracks-logo-256.png 256w,
                 https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/tracks-logo-512.png 512w"
         sizes="(max-width: 768px) 256px, 400px"
         alt="Tracks Logo"
         width="400"
         loading="lazy">
  </picture>
</p>

<p align="center">
  <strong>⚡ Go, fast. A batteries included toolkit for hypermedia servers</strong>
</p>

<p align="center">
  <a href="https://go-tracks.io/">Docs</a> ·
  <a href="#current-status">Status</a> ·
  <a href="#vision">Vision</a> ·
  <a href="#roadmap">Roadmap</a> ·
  <a href="#documentation">Documentation</a> ·
  <a href="#license">License</a>
</p>

---

## Current Status

**Phase 1 (Core Web Layer) is complete!** 🎉

Generated applications now include a production-ready web stack with Chi router, templ templates, HTMX v2, TemplUI components, and comprehensive middleware.

<p align="center">
  <img src="website/static/img/counter-demo.gif" alt="HTMX Counter Demo" width="300">
</p>

**What works now:**

- ✅ Complete web layer
  - Chi router with middleware stack (10 middleware including security headers, CSP, CORS)
  - templ templates with type-safe HTML
  - HTMX v2 with extensions (head-support, idiomorph, response-targets)
  - TemplUI integration with 100+ shadcn-style components
  - TailwindCSS v4 with automatic compilation
  - hashfs content-addressed asset serving with cache headers
  - Air live reload for .templ, .css, and .js files
- ✅ `tracks ui` CLI commands
  - `tracks ui add <components>` - Add TemplUI components
  - `tracks ui list` - List available and installed components
  - `tracks ui upgrade` - Upgrade TemplUI version
- ✅ Project generation (`tracks new` command)
  - Production-ready project scaffolding
  - Choice of database drivers (LibSQL, SQLite3, PostgreSQL)
  - Clean architecture with layered structure (handlers → services → repositories)
  - Working HTMX counter example out of the box
  - Auto-generated `.env` with sensible defaults
  - Cross-platform support (Linux, macOS, Windows)
- ✅ Complete development tooling
  - Makefile with comprehensive targets (dev, test, lint, build, generate-mocks)
  - Docker Compose for all database drivers (auto-started with `make dev`)
  - Air for live reload during development
  - golangci-lint configuration for code quality
  - Mockery integration for automatic mock generation
  - SQLC integration for type-safe database code
  - GitHub Actions CI workflow (tests on Ubuntu, macOS, Windows)
- ✅ Comprehensive documentation
  - Live documentation site with guides and tutorials
  - Middleware configuration guide
  - Caching and asset pipeline documentation
  - CLI reference with examples

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

- Production-ready project structure with clean layered architecture
- Health check endpoint with database connectivity test
- Complete development tooling (Makefile, Air live reload, golangci-lint)
- Docker Compose for local development (auto-started with `make dev`)
- Mockery for automatic test mock generation
- SQLC for type-safe database queries
- Auto-generated `.env` with sensible defaults
- GitHub Actions CI workflow ready to use
- All tests passing out of the box

See the [CLI documentation](https://go-tracks.io/cli/new) and [Getting Started guide](https://go-tracks.io/getting-started/installation) for detailed instructions.

**Coming next:**

- Phase 2: Data Layer - SQLC integration, Goose migrations, repository pattern
- Phase 3: Authentication - Magic links, OTP, OAuth providers
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

### Phase 0: Foundation ✅ Complete

**Goal:** Working CLI that generates project scaffolds

**Status:** Complete · **Released:** v0.1.0

### Phase 1: Core Web Layer ✅ Complete

**Goal:** Complete web stack with routing, templates, and assets

**Key Features:**

- Chi router with 10-middleware stack (security headers, CSP, CORS, compression)
- templ templates with TemplUI components (100+ shadcn-style components)
- HTMX v2 with extensions for dynamic UI
- TailwindCSS v4 with hashfs content-addressed serving
- `tracks ui` CLI for component management

**Status:** Complete · **Released:** v0.3.0

### Future Phases

- **Phase 2 (Current):** Data Layer - SQLC integration, Goose migrations, repository pattern
- **Phase 3:** Authentication - Magic links, OTP, OAuth providers
- **Phase 4:** Interactive TUI - Bubble Tea interface for generators
- **Phase 5:** Production - OpenTelemetry, health checks, deployment

[View Full Roadmap →](./docs/roadmap/README.md)

## Documentation

### Live Documentation

**[📚 View Full Documentation →](https://go-tracks.io/)**

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
make test-e2e-local                  # E2E workflow locally (sqlite3 only)
make test-coverage                   # All tests with coverage reports
```

**Test Types:**

- **Unit Tests** - Colocated with source, run with `-short` flag, use race detector
- **Integration Tests** - File generation, validation, git operations (no external services)
- **E2E Workflow Tests** - Full end-to-end with databases (SQLite3, Postgres, LibSQL) in CI

E2E tests run as shell commands in CI workflows on Ubuntu, testing generated projects with all database drivers. Non-E2E tests run on all platforms.

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

Tracks is in **active development** (Phase 2). The API and generated code structure may change before v1.0.

Generated projects are **suitable for development and prototyping**. Follow development progress via [GitHub Issues](https://github.com/anomalousventures/tracks/issues) and [Roadmap](./docs/roadmap/README.md).

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
- [TemplUI](https://templui.io) - Component library
- [HTMX](https://htmx.org) - HTML-driven interactivity
- [SQLC](https://sqlc.dev/) - SQL code generation
- [hashfs](https://github.com/benbjohnson/hashfs) - Content-addressed serving
- [TailwindCSS](https://tailwindcss.com) - Utility-first CSS
- [Air](https://github.com/cosmtrek/air) - Live reload
- [Goose](https://github.com/pressly/goose) - Database migrations
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework

---

An [Anomalous Venture](https://github.com/enterprises/anomalousventures) by [Aaron Ross](https://github.com/ashmortar)
