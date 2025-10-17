# Tracks Framework - Summary & Overview

**Version:** 1.0.0
**Last Updated:** 2025-01-16

## Quick Links

- [Core Architecture](./1_core_architecture.md) - CLI structure, generated app layout
- [Database Layer](./2_database_layer.md) - SQLC, migrations, repository pattern
- [Authentication](./3_authentication.md) - OTP, OAuth2, session management
- [Authorization & RBAC](./4_authorization_rbac.md) - Casbin-based permissions
- [Web Layer](./5_web_layer.md) - Routing, middleware, handlers
- [Security](./6_security.md) - CSP, headers, protection mechanisms
- [Templates & Assets](./7_templates_assets.md) - templ, i18n, asset pipeline
- [External Services](./8_external_services.md) - Circuit breakers, adapters
- [Configuration](./9_configuration.md) - Environment variables, Viper
- [Background Jobs](./10_background_jobs.md) - Queue adapters, workers
- [Storage](./11_storage.md) - File uploads, S3/R2/local adapters
- [Observability](./12_observability.md) - OpenTelemetry, metrics, logging
- [Testing](./13_testing.md) - Unit, integration, E2E testing
- [Code Generation](./14_code_generation.md) - Interactive generators, TUI
- [MCP Server](./15_mcp_server.md) - AI assistant integration
- [TUI Mode](./16_tui_mode.md) - Interactive dashboard
- [Deployment](./17_deployment.md) - Docker, Kubernetes, production
- [Dependencies](./18_dependencies.md) - Complete dependency list

## Overview

Tracks is a CLI tool and TUI for rapidly generating and developing production-ready Go web applications. It provides Rails-like conventions with Go-idiomatic patterns, focusing on hypermedia-driven applications using HTMX and Alpine.js.

Unlike traditional Go frameworks that use runtime magic, Tracks generates clean, readable Go code that developers would write themselves. All generated code is yours to modify, debug, and deploy without any framework lock-in.

## Core Goals

- **Generate readable, idiomatic Go code with zero magic** - Every line of generated code should be understandable and debuggable
- **Provide batteries-included functionality without framework lock-in** - Start with everything you need, modify anything you want
- **Enable rapid development while maintaining production readiness** - Development speed without sacrificing quality or performance
- **Support modern deployment patterns** - Native support for containers, edge computing, and serverless deployments
- **Database driver flexibility** - Choose between go-libsql (Turso), sqlite3, or PostgreSQL at project creation

## Framework Choices (Built-in, Non-configurable)

These are architectural decisions baked into Tracks. They represent carefully chosen, battle-tested libraries that work well together:

- **Router:** Chi - Lightweight, idiomatic, composable middleware
- **Templates:** templ - Type-safe, compiled templates with Go integration
- **Database Queries:** SQLC - Compile-time SQL validation, zero runtime overhead
- **Configuration Parsing:** Viper - Handles environment variables, config files, and validation
- **Authorization:** Casbin - Flexible RBAC/ABAC with database persistence
- **Testing:** testify - Assertions, mocks, and test suites
- **Sessions:** scs - Secure session management with multiple backends
- **Validation:** go-playground/validator - Struct validation with tags
- **Migrations:** goose - Simple, reliable database migrations
- **Logging:** zerolog - Structured, high-performance logging
- **CLI Framework:** Cobra - Industry-standard CLI building
- **TUI Framework:** Bubble Tea - Interactive terminal UIs

## User Choices (Configurable)

These are implementation choices you make based on your project needs:

### At Project Creation

- **Database Driver:**
  - `go-libsql` (default) - Turso/LibSQL for edge deployments
  - `sqlite3` - Traditional SQLite for single-server apps
  - `postgres` - PostgreSQL for traditional deployments

### Via Configuration

- **Email Provider:**
  - `mailpit` - Local development
  - `ses` - AWS Simple Email Service
  - `sendgrid` - SendGrid API
  - `smtp` - Any SMTP server

- **SMS Provider:**
  - `log` - Development (logs to console)
  - `sns` - AWS Simple Notification Service
  - `twilio` - Twilio Verify API

- **Storage Provider:**
  - `local` - Filesystem storage
  - `s3` - AWS S3
  - `r2` - Cloudflare R2

- **Queue Provider:**
  - `memory` - In-memory for development
  - `sqs` - AWS Simple Queue Service
  - `pubsub` - Google Cloud Pub/Sub

## User Stories

### Development Experience

- As a developer, I want to create a production-ready web app in minutes
- As a developer, I want all generated code to be readable and debuggable
- As a developer, I want to choose my database based on deployment needs
- As a developer, I want compile-time validation for SQL queries and templates
- As a developer, I want hot-reload during development

### Team & Operations

- As a team lead, I want consistent project structure across all our Go apps
- As a DevOps engineer, I want applications that are easy to deploy and monitor
- As a DevOps engineer, I want built-in health checks and metrics endpoints
- As a security engineer, I want secure defaults and audit trails

### Application Features

- As a user, I want fast page loads with minimal JavaScript
- As a user, I want secure authentication without passwords
- As a user, I want the app to work on all devices
- As a product owner, I want flexible permission management

## Key Features

### Developer Experience

- Zero-config local development with sensible defaults
- Interactive TUI for code generation and monitoring
- Hot-reload with Air for instant feedback
- Type-safe everything: routes, queries, templates

### Production Ready

- Database migrations with up/down support
- Circuit breakers for external services
- Structured logging with correlation IDs
- OpenTelemetry instrumentation
- Graceful shutdown handling
- Health check endpoints

### Security First

- Secure session management
- CSRF protection
- Content Security Policy with nonces
- Rate limiting
- SQL injection prevention via SQLC
- XSS protection via template escaping

### Modern Architecture

- Hypermedia-driven (HTMX + Alpine.js)
- Progressive enhancement
- Server-side rendering
- Minimal client JavaScript
- WebSocket support (future)

## Quick Start

```bash
# Install Tracks
go install github.com/yourusername/tracks@latest

# Create a new application
tracks new myapp --db postgres

# Start development server
cd myapp
tracks dev

# Generate a resource
tracks generate resource post title:string content:text author:relation

# Run migrations
tracks db migrate

# Run tests
tracks test

# Build for production
tracks build
```

## Project Philosophy

1. **Clarity over Cleverness** - Code should be obvious, not clever
2. **Explicit over Implicit** - No hidden magic or conventions
3. **Composition over Inheritance** - Use interfaces and composition
4. **Errors are Values** - Handle errors explicitly
5. **Documentation as Code** - Generate docs from code

## What Tracks is NOT

- **Not a runtime framework** - Tracks generates code, it doesn't run in production
- **Not a microframework** - Tracks is batteries-included by design
- **Not a REST API framework** - Tracks focuses on hypermedia-driven applications
- **Not a CMS** - Tracks generates custom applications, not content management systems

## Getting Help

- [Installation Guide](./1_core_architecture.md#installation)
- [Tutorial: Building a Blog](../tutorials/blog.md)
- [Example Applications](../examples/)
- [Troubleshooting Guide](../troubleshooting.md)
- [GitHub Issues](https://github.com/yourusername/tracks/issues)

## Contributing

Tracks is open source and welcomes contributions. See our [Contributing Guide](../CONTRIBUTING.md) for details on:

- Setting up the development environment
- Running the test suite
- Submitting pull requests
- Code style guidelines

## License

MIT License - see [LICENSE](../LICENSE) for details.

---

**Navigation:** [Next: Core Architecture →](./1_core_architecture.md)
