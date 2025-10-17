# Dependencies

**[← Back to Summary](./0_summary.md)**

## Overview

This document lists all Go dependencies required by the Tracks framework, organized by category. Each dependency is chosen for specific reasons: performance, reliability, active maintenance, and alignment with Go best practices.

## Complete go.mod

```go
module github.com/user/tracks

go 1.23

require (
    // Core Web Framework
    github.com/go-chi/chi/v5 v5.0.10
    github.com/go-chi/httprate v0.8.0
    github.com/go-chi/cors v1.2.1

    // Templates and UI
    github.com/a-h/templ v0.2.513
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.17.1
    github.com/charmbracelet/lipgloss v0.9.1

    // Database
    github.com/tursodatabase/libsql-client-go v0.0.0-20240416075003-747366ff79c4
    github.com/lib/pq v1.10.9
    github.com/mattn/go-sqlite3 v1.14.19
    github.com/pressly/goose/v3 v3.17.0
    github.com/XSAM/otelsql v0.27.0

    // UUID (with UUIDv7 support)
    github.com/gofrs/uuid/v5 v5.0.0

    // Configuration
    github.com/spf13/viper v1.18.2
    github.com/spf13/cobra v1.8.0
    github.com/joho/godotenv v1.5.1

    // Session Management
    github.com/alexedwards/scs/v2 v2.7.0
    github.com/alexedwards/scs/redisstore v0.0.0-20231113091146-cef8b05e1a82

    // Authorization
    github.com/casbin/casbin/v2 v2.82.0
    github.com/casbin/gorm-adapter/v3 v3.20.0

    // Authentication
    github.com/markbates/goth v1.78.0
    github.com/gorilla/sessions v1.2.2

    // Validation and Security
    github.com/go-playground/validator/v10 v10.16.0
    github.com/microcosm-cc/bluemonday v1.0.26
    github.com/unrolled/secure v1.13.0

    // Logging
    github.com/rs/zerolog v1.31.0

    // Observability
    go.opentelemetry.io/otel v1.21.0
    go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.44.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
    go.opentelemetry.io/otel/sdk v1.21.0
    go.opentelemetry.io/otel/sdk/metric v1.21.0
    github.com/riandyrn/otelchi v0.5.1
    github.com/prometheus/client_golang v1.18.0

    // AWS SDK (for S3, SES, SNS, SQS)
    github.com/aws/aws-sdk-go-v2 v1.24.1
    github.com/aws/aws-sdk-go-v2/config v1.26.6
    github.com/aws/aws-sdk-go-v2/service/s3 v1.47.8
    github.com/aws/aws-sdk-go-v2/service/ses v1.19.6
    github.com/aws/aws-sdk-go-v2/service/sns v1.26.7
    github.com/aws/aws-sdk-go-v2/service/sqs v1.29.7

    // External Services
    github.com/sendgrid/sendgrid-go v3.14.0+incompatible
    github.com/twilio/twilio-go v1.16.0
    github.com/stripe/stripe-go/v76 v76.8.0

    // Circuit Breaker
    github.com/sony/gobreaker v0.5.0

    // Utilities
    github.com/jaevor/go-nanoid v1.3.0
    github.com/benbjohnson/hashfs v0.2.1
    github.com/dustin/go-humanize v1.0.1
    golang.org/x/text v0.14.0
    github.com/invopop/ctxi18n v0.8.1

    // Testing
    github.com/stretchr/testify v1.8.4
    github.com/stretchr/mock v1.8.4
    github.com/jaswdr/faker v1.19.1

    // Development Tools
    github.com/cosmtrek/air v1.49.0
    github.com/golangci/golangci-lint v1.55.2
)

require (
    // Indirect dependencies (auto-managed)
    github.com/fsnotify/fsnotify v1.7.0 // indirect
    github.com/hashicorp/hcl v1.0.0 // indirect
    github.com/magiconair/properties v1.8.7 // indirect
    github.com/mitchellh/mapstructure v1.5.0 // indirect
    github.com/pelletier/go-toml/v2 v2.1.1 // indirect
    github.com/spf13/afero v1.11.0 // indirect
    github.com/spf13/cast v1.6.0 // indirect
    github.com/spf13/jwalterweatherman v1.1.0 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
    github.com/subosito/gotenv v1.6.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

## Dependency Categories

### Core Framework

#### Chi Router

```go
github.com/go-chi/chi/v5 v5.0.10
```

- **Purpose**: HTTP router and middleware framework
- **Why Chi**: Lightweight, idiomatic Go, compatible with net/http, excellent middleware support
- **Alternative considered**: Gin, Echo (both less idiomatic)

#### Templ

```go
github.com/a-h/templ v0.2.513
```

- **Purpose**: Type-safe HTML templating
- **Why Templ**: Compile-time checking, Go integration, performance
- **Alternative considered**: html/template (less type safety)

### Database

#### LibSQL Client

```go
github.com/tursodatabase/libsql-client-go v0.0.0-20240416075003-747366ff79c4
```

- **Purpose**: Turso/LibSQL database driver
- **Why LibSQL**: Edge-ready, SQLite compatible, built for distributed apps
- **Note**: Default driver for Tracks

#### PostgreSQL Driver

```go
github.com/lib/pq v1.10.9
```

- **Purpose**: PostgreSQL database driver
- **Why lib/pq**: Pure Go, well-maintained, production-tested
- **Alternative**: pgx (more features but more complex)

#### SQLite Driver

```go
github.com/mattn/go-sqlite3 v1.14.19
```

- **Purpose**: Alternative SQLite driver
- **Why go-sqlite3**: Most mature SQLite driver, full feature support
- **Note**: Requires CGO

#### Goose Migrations

```go
github.com/pressly/goose/v3 v3.17.0
```

- **Purpose**: Database migration tool
- **Why Goose**: Simple, supports multiple databases, Go-based migrations
- **Alternative considered**: golang-migrate (more complex)

### UUID Generation

#### gofrs/uuid

```go
github.com/gofrs/uuid/v5 v5.0.0
```

- **Purpose**: UUID generation with UUIDv7 support
- **Why gofrs**: UUIDv7 support (timestamp-ordered), drop-in replacement
- **Critical**: DO NOT use google/uuid (no UUIDv7 support)

### Configuration

#### Viper

```go
github.com/spf13/viper v1.18.2
```

- **Purpose**: Configuration management
- **Why Viper**: Multiple format support, env var binding, hot reload
- **Note**: Used for parsing only, not a framework requirement

#### Cobra

```go
github.com/spf13/cobra v1.8.0
```

- **Purpose**: CLI framework
- **Why Cobra**: Industry standard, excellent documentation, subcommands
- **Alternative considered**: urfave/cli (less features)

### Session Management

#### SCS

```go
github.com/alexedwards/scs/v2 v2.7.0
```

- **Purpose**: HTTP session management
- **Why SCS**: Secure by default, multiple store backends, middleware ready
- **Alternative considered**: gorilla/sessions (less secure defaults)

### Authorization

#### Casbin

```go
github.com/casbin/casbin/v2 v2.82.0
```

- **Purpose**: Authorization library
- **Why Casbin**: RBAC/ABAC support, flexible policies, well-documented
- **Alternative considered**: Open Policy Agent (more complex)

### Authentication

#### Goth

```go
github.com/markbates/goth v1.78.0
```

- **Purpose**: Multi-provider OAuth authentication
- **Why Goth**: Supports 50+ providers, easy integration, active maintenance
- **Alternative considered**: golang/oauth2 (manual provider setup)

### Validation

#### Validator

```go
github.com/go-playground/validator/v10 v10.16.0
```

- **Purpose**: Struct and field validation
- **Why Validator**: Extensive validation tags, custom validators, i18n support
- **Alternative considered**: ozzo-validation (less community)

#### Bluemonday

```go
github.com/microcosm-cc/bluemonday v1.0.26
```

- **Purpose**: HTML sanitization
- **Why Bluemonday**: XSS prevention, configurable policies, well-tested
- **Critical**: Essential for user-generated content

### Logging

#### Zerolog

```go
github.com/rs/zerolog v1.31.0
```

- **Purpose**: Structured JSON logging
- **Why Zerolog**: Zero allocation, high performance, clean API
- **Alternative considered**: zap (more complex API)

### Observability

#### OpenTelemetry

```go
go.opentelemetry.io/otel v1.21.0
go.opentelemetry.io/otel/sdk v1.21.0
```

- **Purpose**: Distributed tracing and metrics
- **Why OpenTelemetry**: Vendor-agnostic, industry standard, comprehensive
- **Alternative considered**: Datadog APM (vendor lock-in)

#### Prometheus Client

```go
github.com/prometheus/client_golang v1.18.0
```

- **Purpose**: Metrics exposition
- **Why Prometheus**: Industry standard, excellent ecosystem, time-series focus
- **Used with**: OpenTelemetry for unified observability

### Cloud Services

#### AWS SDK v2

```go
github.com/aws/aws-sdk-go-v2 v1.24.1
```

- **Purpose**: AWS service integration
- **Why SDK v2**: Modern API, better performance, context support
- **Services used**: S3 (storage), SES (email), SNS (SMS), SQS (queues)

### External Services

#### SendGrid

```go
github.com/sendgrid/sendgrid-go v3.14.0+incompatible
```

- **Purpose**: Transactional email
- **Why SendGrid**: Reliable delivery, good analytics, template support

#### Twilio

```go
github.com/twilio/twilio-go v1.16.0
```

- **Purpose**: SMS messaging
- **Why Twilio**: Market leader, global coverage, reliable

### Circuit Breaker

#### Sony Gobreaker

```go
github.com/sony/gobreaker v0.5.0
```

- **Purpose**: Circuit breaker pattern implementation
- **Why Gobreaker**: Simple API, battle-tested, follows the pattern correctly
- **Used for**: External service resilience

### TUI

#### Bubble Tea

```go
github.com/charmbracelet/bubbletea v0.25.0
github.com/charmbracelet/bubbles v0.17.1
github.com/charmbracelet/lipgloss v0.9.1
```

- **Purpose**: Terminal UI framework
- **Why Bubble Tea**: Elm architecture, excellent components, active development
- **Alternative considered**: tview (less modern architecture)

### Utilities

#### NanoID

```go
github.com/jaevor/go-nanoid v1.3.0
```

- **Purpose**: URL-safe unique ID generation
- **Why NanoID**: Shorter than UUID, customizable alphabet, secure
- **Used for**: User-facing IDs, tokens

#### HashFS

```go
github.com/benbjohnson/hashfs v0.2.1
```

- **Purpose**: Content-addressed file embedding
- **Why HashFS**: Cache busting, efficient static file serving
- **Used for**: Asset pipeline

### Testing

#### Testify

```go
github.com/stretchr/testify v1.8.4
```

- **Purpose**: Testing toolkit
- **Why Testify**: Assertions, mocks, suites, widely adopted
- **Alternative considered**: Standard library only (more verbose)

#### Faker

```go
github.com/jaswdr/faker v1.19.1
```

- **Purpose**: Fake data generation
- **Why Faker**: Comprehensive data types, deterministic option
- **Used for**: Test fixtures

### Development

#### Air

```go
github.com/cosmtrek/air v1.49.0
```

- **Purpose**: Live reload for development
- **Why Air**: Fast, configurable, supports complex projects
- **Alternative considered**: realize (less maintained)

## Version Policy

### Semantic Versioning

- All direct dependencies use semantic versioning
- Pin to minor versions for stability
- Review major version updates carefully

### Update Strategy

1. **Security updates**: Apply immediately
2. **Patch updates**: Apply monthly
3. **Minor updates**: Test in staging first
4. **Major updates**: Full regression testing required

### Dependency Audit

```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Check for updates
go list -u -m all

# Update all dependencies
go get -u ./...
go mod tidy

# Verify dependencies
go mod verify
```

## License Compatibility

All dependencies are checked for license compatibility:

- **MIT**: Most dependencies (Chi, Templ, Zerolog, etc.)
- **Apache 2.0**: OpenTelemetry, AWS SDK, Casbin
- **BSD**: PostgreSQL driver, UUID library
- **MPL-2.0**: HashiCorp libraries (transitive)

## Minimal Dependencies Principle

Tracks follows these principles for dependencies:

1. **Prefer standard library** when adequate
2. **One tool per job** - avoid duplicate functionality
3. **Active maintenance** - no abandoned projects
4. **Community adoption** - prefer widely-used libraries
5. **Security first** - regular vulnerability scanning

## Dependency Graph

```text
tracks
├── Web Layer
│   ├── chi (router)
│   ├── templ (templates)
│   └── scs (sessions)
├── Data Layer
│   ├── libsql-client-go (primary DB)
│   ├── goose (migrations)
│   └── gofrs/uuid (IDs)
├── Business Logic
│   ├── casbin (authorization)
│   ├── validator (validation)
│   └── gobreaker (resilience)
├── Infrastructure
│   ├── viper (config)
│   ├── zerolog (logging)
│   └── opentelemetry (observability)
└── External Services
    ├── aws-sdk-go-v2 (cloud)
    ├── sendgrid-go (email)
    └── twilio-go (SMS)
```

## Next Steps

- Back to [← Deployment](./17_deployment.md)
- Return to [Summary](./0_summary.md)

---

**Congratulations!** You've completed reviewing all Tracks framework documentation. The framework is now fully documented with modular, LLM-friendly files that can be easily consumed and understood.
