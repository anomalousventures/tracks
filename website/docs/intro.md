# Introduction

Welcome to **Tracks** - a Rails-like web framework for Go that generates idiomatic, production-ready applications.

## What is Tracks?

Tracks is a command-line tool that generates and manages Go web applications with the productivity of Rails and the performance of Go. It provides a complete, opinionated framework with code generation, type-safe templates, and modern tooling—all producing idiomatic Go code that you'd write yourself.

## Key Features

### 🚀 Rapid Development

- **Interactive TUI** - Beautiful terminal interface for project setup and code generation
- **Smart Generators** - Create complete CRUD resources with a single command
- **Live Reload** - Hot reload in development with Air
- **MCP Server** - AI-powered development via Model Context Protocol

### 🏗️ Modern Architecture

- **Type-Safe Templates** - [templ](https://github.com/a-h/templ) for compile-time HTML safety
- **Type-Safe SQL** - [SQLC](https://sqlc.dev/) generates Go code from SQL queries
- **Chi Router** - Lightweight, idiomatic HTTP routing
- **Dependency Injection** - Services built for easy testing

### 🔐 Security First

- **Built-in Authentication** - Magic link, OTP, and OAuth providers
- **RBAC Authorization** - Casbin-powered role-based access control
- **Security Headers** - CSP, HSTS, and more configured by default
- **Input Validation** - Comprehensive validation with go-playground/validator

## Philosophy

### Idiomatic Go

Tracks generates code that looks like it was hand-written by an experienced Go developer:

- No magic or reflection in generated code
- Standard library patterns
- Clear error handling
- Interface-based design for testing

### Rails-Inspired Productivity

- Convention over configuration
- Generators for rapid development
- Built-in authentication and authorization
- Asset pipeline and view system

### Production Ready

- OpenTelemetry observability
- Database connection pooling
- Graceful shutdown
- Health check endpoints
- Structured logging

## Project Status

Tracks is currently in **pre-release development** (v0.x). The API may change before v1.0.

## Next Steps

Ready to get started? Check out our [Quick Start Guide](./getting-started/quick-start.md) or [Installation Instructions](./getting-started/installation.md).

## Community

- **GitHub**: [anomalousventures/tracks](https://github.com/anomalousventures/tracks)
- **Discussions**: [GitHub Discussions](https://github.com/anomalousventures/tracks/discussions)
- **Issues**: [Report bugs or request features](https://github.com/anomalousventures/tracks/issues)
