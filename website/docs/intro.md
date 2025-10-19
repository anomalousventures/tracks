# Introduction

Welcome to **Tracks** - a code-generating web framework for Go that produces idiomatic, production-ready applications.

:::info Current Status

Tracks is in **Phase 0 (Foundation)** development. The CLI tool and project scaffolding are being built.

**What works now:**

- ✅ Project structure and monorepo setup
- ✅ Documentation and roadmap
- 🚧 CLI infrastructure (in progress)

**Coming next:** `tracks new` command, template engine, project generation tooling.

See the [Roadmap](https://github.com/anomalousventures/tracks/blob/main/docs/roadmap/README.md) for details.

:::

## What is Tracks?

Tracks will be a command-line tool that generates and manages Go web applications. Built for developers who want the productivity of modern frameworks with the performance and simplicity of Go.

The framework generates idiomatic Go code - the kind you'd write yourself. No magic, no reflection, just clean, testable, production-ready applications.

## Design Principles

### Type-Safe Everything

- **[templ](https://github.com/a-h/templ)** for compile-time HTML safety
- **[SQLC](https://sqlc.dev/)** generates type-safe Go from SQL
- No runtime reflection in generated code

### Hypermedia-First

- **HTMX** integration for dynamic UIs without JavaScript
- Server-rendered templates with progressive enhancement
- RESTful patterns with HTML as the engine of application state

### Idiomatic Go

- Generated code looks hand-written by an experienced Go developer
- Standard library patterns and clear error handling
- Dependency injection via interfaces for easy testing

### Batteries Included

- Authentication (magic links, OTP, OAuth)
- Authorization (Casbin RBAC)
- Observability (OpenTelemetry)
- Development tooling (hot reload, linting, CI/CD templates)

## Roadmap

Tracks development is organized into 7 phases:

- **Phase 0 (Current):** Foundation - CLI tool and project scaffolding
- **Phase 1:** Core Web Layer - Chi router, handlers, middleware, templ templates
- **Phase 2:** Database Layer - SQLC, Goose migrations, LibSQL/PostgreSQL support
- **Phase 3:** Authentication - Magic links, OTP, OAuth providers
- **Phase 4:** Interactive TUI - Bubble Tea interface for generators
- **Phase 5:** Production - OpenTelemetry, health checks, deployment
- **Phase 6:** Authorization - Casbin RBAC
- **Phase 7:** Advanced Features - Real-time, file uploads, background jobs

[View Full Roadmap →](https://github.com/anomalousventures/tracks/blob/main/docs/roadmap/README.md)

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

## Project Status

Tracks is in **pre-alpha development** (Phase 0). The API and generated code structure may change significantly before v1.0.

**Not ready for production use.** Follow development progress via [GitHub Issues](https://github.com/anomalousventures/tracks/issues) and the [Roadmap](https://github.com/anomalousventures/tracks/blob/main/docs/roadmap/README.md).

## Community

- **GitHub**: [anomalousventures/tracks](https://github.com/anomalousventures/tracks)
- **Discussions**: [GitHub Discussions](https://github.com/anomalousventures/tracks/discussions)
- **Issues**: [Report bugs or request features](https://github.com/anomalousventures/tracks/issues)

## Next Steps

Once Phase 0 is complete, you'll be able to:

1. Install Tracks CLI
2. Run `tracks new myapp` to generate a project
3. Explore the generated code structure
4. Read the [Architecture Guide](./core/architecture.md) (coming soon)

For now, check out the [Roadmap](https://github.com/anomalousventures/tracks/blob/main/docs/roadmap/README.md) to see what's being built.
