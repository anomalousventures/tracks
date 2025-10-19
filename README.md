# Tracks

<p align="center">
  <img src="https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/tracks-logo.svg" alt="Tracks Logo" width="400">
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
- 🚧 CLI infrastructure (in progress)

**Coming next:**

- Phase 0: CLI tool with `tracks new` command
- Phase 1: Core web layer (routing, handlers, middleware)
- Phase 2: Database layer (SQLC, migrations)
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
- RESTful patterns with HTML as the engine of application state

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

- [CLAUDE.md](./CLAUDE.md) - Development commands and project structure
- [Roadmap](./docs/roadmap/README.md) - Phase breakdown and epic details
- [GitHub Issues](https://github.com/anomalousventures/tracks/issues) - Task tracking

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
- Standard HTTP and REST patterns

### Convention Over Configuration

Sensible defaults with escape hatches:

- Standard project structure and file organization
- Generators follow consistent patterns
- Override defaults when needed
- Configuration via environment variables and YAML

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

## Project Status

Tracks is in **pre-alpha development** (Phase 0). The API and generated code structure may change significantly before v1.0.

**Not ready for production use.** Follow development progress via [GitHub Issues](https://github.com/anomalousventures/tracks/issues) and [Roadmap](./docs/roadmap/README.md).

## Contributing

Contributions are welcome! The project is in active early development.

**Ways to contribute:**

- Report bugs and request features via [GitHub Issues](https://github.com/anomalousventures/tracks/issues)
- Discuss ideas in [GitHub Discussions](https://github.com/anomalousventures/tracks/discussions)
- Submit pull requests (see [CLAUDE.md](./CLAUDE.md) for development setup)

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
