# Configuration

**[← Back to Summary](./0_summary.md)**

## Overview

Configuration management in Tracks uses **Viper** for parsing and validation (a framework choice). The system supports both development (.env files) and production (environment variables) workflows with clear distinction between framework choices and user-configurable options.

**Important:** This section describes configuration for the **generated application**, not the Tracks CLI/TUI.

For Tracks CLI project metadata (`.tracks.yaml`), see [ADR-007](../adr/007-configuration-file-separation.md). Generated applications use `.env` for runtime configuration and do NOT read `.tracks.yaml`.

## Goals

- Viper for all configuration parsing and validation (Framework Choice)
- Development uses .env files, production uses environment variables
- Runtime validation with clear error messages
- Sensible defaults that work without configuration
- Clear separation between framework and user choices

## User Stories

- As a developer, I want zero-config local development with sensible defaults
- As a DevOps engineer, I want configuration via environment variables in production
- As a developer, I want clear errors for missing required configuration
- As a security engineer, I want secrets never committed to git (.env in .gitignore)
- As a developer, I want Viper to handle all config complexity

## Framework Choices vs User Choices

### Framework Choices (Built-in, Non-configurable)

These are architectural decisions baked into Tracks. They cannot be changed after project generation:

- **Viper** - Configuration parsing and validation
- **Chi** - HTTP router
- **templ** - Template engine
- **SQLC** - SQL code generation
- **Casbin** - RBAC authorization
- **testify** - Testing framework
- **scs** - Session management
- **goose** - Database migrations
- **zerolog** - Structured logging

### User Choices (Configurable)

These can be configured for different environments and deployments:

#### At Project Creation

- **Database driver** (selected once, cannot change):
  - `go-libsql` (default) - Turso/LibSQL for edge deployments
  - `sqlite3` - Traditional SQLite for single-server apps
  - `postgres` - PostgreSQL for traditional deployments

#### Via Configuration

- **Email provider**:
  - `mailpit` - Local development
  - `ses` - AWS Simple Email Service
  - `sendgrid` - SendGrid API
  - `smtp` - Any SMTP server
  - `log` - Log to console (development)

- **SMS provider**:
  - `log` - Log to console (development)
  - `sns` - AWS Simple Notification Service
  - `twilio` - Twilio Verify API

- **Storage provider**:
  - `local` - Filesystem storage
  - `s3` - AWS S3
  - `r2` - Cloudflare R2

- **Queue provider**:
  - `memory` - In-memory (development)
  - `sqs` - AWS Simple Queue Service
  - `pubsub` - Google Cloud Pub/Sub

## Configuration Structure

```go
// internal/config/config.go
package config

import (
    "fmt"
    "github.com/spf13/viper"
)

type Config struct {
    // Required
    DatabaseURL    string `mapstructure:"DATABASE_URL"`
    DatabaseDriver string `mapstructure:"DATABASE_DRIVER"` // go-libsql, sqlite3, or postgres
    SessionKey     string `mapstructure:"SESSION_KEY"`

    // Optional with defaults
    Port        int    `mapstructure:"PORT"`
    Environment string `mapstructure:"APP_ENV"`
    LogLevel    string `mapstructure:"LOG_LEVEL"`

    // External services (optional)
    Email       EmailConfig
    SMS         SMSConfig
    Storage     StorageConfig
    Queue       QueueConfig
    RateLimit   RateLimitConfig
}

type EmailConfig struct {
    Provider string `mapstructure:"EMAIL_PROVIDER"`
    From     string `mapstructure:"EMAIL_FROM"`

    // AWS SES
    SESRegion    string `mapstructure:"AWS_REGION"`
    SESAccessKey string `mapstructure:"AWS_ACCESS_KEY_ID"`
    SESSecretKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`

    // Mailpit
    MailpitHost string `mapstructure:"MAILPIT_HOST"`
    MailpitPort int    `mapstructure:"MAILPIT_PORT"`

    // SendGrid
    SendGridAPIKey string `mapstructure:"SENDGRID_API_KEY"`

    // SMTP
    SMTPHost     string `mapstructure:"SMTP_HOST"`
    SMTPPort     int    `mapstructure:"SMTP_PORT"`
    SMTPUsername string `mapstructure:"SMTP_USERNAME"`
    SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
}

type SMSConfig struct {
    Provider string `mapstructure:"SMS_PROVIDER"`

    // AWS SNS
    SNSRegion    string `mapstructure:"AWS_REGION"`
    SNSAccessKey string `mapstructure:"AWS_ACCESS_KEY_ID"`
    SNSSecretKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`

    // Twilio
    TwilioAccountSID string `mapstructure:"TWILIO_ACCOUNT_SID"`
    TwilioAuthToken  string `mapstructure:"TWILIO_AUTH_TOKEN"`
    TwilioServiceID  string `mapstructure:"TWILIO_SERVICE_ID"`
    TwilioFromNumber string `mapstructure:"TWILIO_FROM_NUMBER"`
}

type StorageConfig struct {
    Provider string `mapstructure:"STORAGE_PROVIDER"`

    // S3/R2
    S3Region    string `mapstructure:"AWS_REGION"`
    S3Bucket    string `mapstructure:"S3_BUCKET"`
    S3AccessKey string `mapstructure:"AWS_ACCESS_KEY_ID"`
    S3SecretKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
    S3Endpoint  string `mapstructure:"S3_ENDPOINT"` // For R2 or MinIO

    // Local
    LocalPath string `mapstructure:"STORAGE_PATH"`
}

type QueueConfig struct {
    Provider string `mapstructure:"QUEUE_PROVIDER"`
    URL      string `mapstructure:"QUEUE_URL"`
}

type RateLimitConfig struct {
    OTPPerHour     int `mapstructure:"RATE_LIMIT_OTP"`
    LoginPerMinute int `mapstructure:"RATE_LIMIT_LOGIN"`
}
```

## Loading Configuration with Viper

```go
// internal/config/loader.go
package config

import (
    "fmt"
    "github.com/spf13/viper"
)

// Load configuration using Viper
func Load() (*Config, error) {
    v := viper.New()

    // Set defaults
    v.SetDefault("PORT", 8080)
    v.SetDefault("APP_ENV", "production")
    v.SetDefault("LOG_LEVEL", "info")
    v.SetDefault("DATABASE_DRIVER", "go-libsql") // Default driver
    v.SetDefault("EMAIL_PROVIDER", "log")
    v.SetDefault("EMAIL_FROM", "noreply@example.com")
    v.SetDefault("SMS_PROVIDER", "log")
    v.SetDefault("STORAGE_PROVIDER", "local")
    v.SetDefault("STORAGE_PATH", "./uploads")
    v.SetDefault("QUEUE_PROVIDER", "memory")
    v.SetDefault("RATE_LIMIT_OTP", 5)
    v.SetDefault("RATE_LIMIT_LOGIN", 10)
    v.SetDefault("MAILPIT_HOST", "localhost")
    v.SetDefault("MAILPIT_PORT", 1025)

    // Read from .env file (development only, gitignored)
    // Note: Do NOT read .tracks.yaml - that's for CLI metadata only
    v.SetConfigFile(".env")
    v.SetConfigType("env")
    _ = v.ReadInConfig() // Ignore error if file doesn't exist

    // Environment variables override everything (production)
    v.AutomaticEnv()

    // Unmarshal into config struct
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func (c *Config) Validate() error {
    // Validate required fields
    if c.DatabaseURL == "" {
        return fmt.Errorf("DATABASE_URL is required")
    }

    if len(c.SessionKey) < 32 {
        return fmt.Errorf("SESSION_KEY must be at least 32 characters")
    }

    // Validate database driver
    switch c.DatabaseDriver {
    case "go-libsql", "sqlite3", "postgres":
        // Valid drivers
    default:
        return fmt.Errorf("invalid DATABASE_DRIVER: %s (must be go-libsql, sqlite3, or postgres)", c.DatabaseDriver)
    }

    // Validate provider-specific requirements
    switch c.Queue.Provider {
    case "sqs":
        if c.Queue.URL == "" {
            return fmt.Errorf("QUEUE_URL required for SQS")
        }
    case "pubsub":
        if c.Queue.URL == "" {
            return fmt.Errorf("QUEUE_URL (project ID) required for Pub/Sub")
        }
    case "memory":
        // No validation needed
    default:
        return fmt.Errorf("unknown queue provider: %s", c.Queue.Provider)
    }

    switch c.Email.Provider {
    case "ses":
        if c.Email.SESAccessKey == "" {
            return fmt.Errorf("AWS_ACCESS_KEY_ID required for SES")
        }
    case "sendgrid":
        if c.Email.SendGridAPIKey == "" {
            return fmt.Errorf("SENDGRID_API_KEY required for SendGrid")
        }
    case "smtp":
        if c.Email.SMTPHost == "" {
            return fmt.Errorf("SMTP_HOST required for SMTP")
        }
    case "mailpit", "log":
        // No validation needed
    default:
        return fmt.Errorf("unknown email provider: %s", c.Email.Provider)
    }

    return nil
}
```

## Development Configuration

```bash
# ============================================================================
# .env - Application Runtime Configuration (development only)
# DO NOT commit this file - add to .gitignore
# For CLI project metadata, see .tracks.yaml (committed to git)
# ============================================================================

# Required
# Database URL format depends on driver:
#   go-libsql: file:dev.db or libsql://...
#   sqlite3: dev.db
#   postgres: postgresql://user:pass@localhost/dbname
DATABASE_URL=file:dev.db
DATABASE_DRIVER=go-libsql  # Options: go-libsql, sqlite3, postgres
SESSION_KEY=dev-session-key-change-in-production-min-32-chars

# Required for go-libsql or sqlite3 drivers
CGO_ENABLED=1

# Optional - defaults shown
PORT=8080
APP_ENV=development
LOG_LEVEL=debug

# Email (development)
EMAIL_PROVIDER=mailpit
EMAIL_FROM=dev@localhost
MAILPIT_HOST=localhost
MAILPIT_PORT=1025

# Storage (development)
STORAGE_PROVIDER=local
STORAGE_PATH=./uploads

# Queue (development)
QUEUE_PROVIDER=memory

# Rate limiting (relaxed for dev)
RATE_LIMIT_OTP=100
RATE_LIMIT_LOGIN=100
```

## Production Configuration

Production uses environment variables only (no .env file deployed):

```bash
# Minimal production configuration (go-libsql with Turso)
export DATABASE_URL="libsql://db.turso.io?authToken=..."
export DATABASE_DRIVER="go-libsql"
export SESSION_KEY="randomly-generated-32-char-minimum-key"
export CGO_ENABLED=1  # Required for go-libsql

# Alternative: Using PostgreSQL
export DATABASE_URL="postgresql://user:pass@host/db"
export DATABASE_DRIVER="postgres"
export SESSION_KEY="randomly-generated-32-char-minimum-key"
# CGO_ENABLED not required for PostgreSQL

# With AWS services
export DATABASE_URL="libsql://db.turso.io?authToken=..."
export DATABASE_DRIVER="go-libsql"
export CGO_ENABLED=1
export SESSION_KEY="randomly-generated-32-char-minimum-key"
export AWS_REGION="us-east-1"
export AWS_ACCESS_KEY_ID="AKIA..."
export AWS_SECRET_ACCESS_KEY="..."
export QUEUE_PROVIDER="sqs"
export QUEUE_URL="https://sqs.us-east-1.amazonaws.com/123456/myqueue"
export EMAIL_PROVIDER="ses"
export STORAGE_PROVIDER="s3"
export S3_BUCKET="my-app-uploads"

# With rate limit adjustments
export RATE_LIMIT_OTP="3"        # Stricter in production
export RATE_LIMIT_LOGIN="5"
```

## Validation & Error Handling

### Goals

- Consistent validation across all layers
- User-friendly error messages
- Type-safe validation with struct tags
- HTMX-compatible error responses
- Toast notifications for success/error states

### DTO Validation

```go
// internal/domain/users/dto.go
package users

type CreateUserDTO struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserDTO struct {
    Email string `json:"email" validate:"omitempty,email"`
    Name  string `json:"name" validate:"omitempty,min=2,max=100"`
}
```

### Validation in Handlers

```go
// internal/http/handlers/user_handler.go
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var dto users.CreateUserDTO
    dto.Email = r.FormValue("email")
    dto.Name = r.FormValue("name")
    dto.Password = r.FormValue("password")

    if err := validator.Struct(dto); err != nil {
        w.WriteHeader(http.StatusUnprocessableEntity)
        views.UserForm(dto, validationErrors(err)).Render(r.Context(), w)
        return
    }

    user, err := h.svc.Create(r.Context(), dto)
    if err != nil {
        handleServiceError(w, r, err)
        return
    }

    // HTMX toast trigger
    w.Header().Set("HX-Trigger", `{"toast":{"type":"success","msg":"User created"}}`)
    views.UserRow(user).Render(r.Context(), w)
}

func validationErrors(err error) map[string]string {
    errors := make(map[string]string)
    for _, err := range err.(validator.ValidationErrors) {
        errors[err.Field()] = friendlyMessage(err)
    }
    return errors
}

func friendlyMessage(err validator.FieldError) string {
    field := err.Field()
    tag := err.Tag()

    switch tag {
    case "required":
        return fmt.Sprintf("%s is required", field)
    case "email":
        return "Please enter a valid email address"
    case "min":
        return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
    case "max":
        return fmt.Sprintf("%s cannot exceed %s characters", field, err.Param())
    default:
        return fmt.Sprintf("%s is invalid", field)
    }
}
```

## Environment-Specific Behavior

```go
// internal/app/app.go
func NewApp(cfg *Config) (*App, error) {
    app := &App{
        config: cfg,
    }

    // Environment-specific setup
    switch cfg.Environment {
    case "development":
        app.setupDevelopment()
    case "production":
        app.setupProduction()
    case "test":
        app.setupTest()
    default:
        return nil, fmt.Errorf("unknown environment: %s", cfg.Environment)
    }

    return app, nil
}

func (a *App) setupDevelopment() {
    // Verbose logging
    zerolog.SetGlobalLevel(zerolog.DebugLevel)

    // Pretty console output
    log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

    // Relaxed security for development
    a.csrfExempt = true
    a.allowHTTP = true
}

func (a *App) setupProduction() {
    // JSON logging
    zerolog.SetGlobalLevel(zerolog.InfoLevel)

    // Structured logging
    log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

    // Strict security
    a.csrfExempt = false
    a.allowHTTP = false
}
```

## Best Practices

1. **Never commit .env files** - Use .env.example as documentation
2. **Use strong session keys** - Minimum 32 characters, randomly generated
3. **Validate early** - Check configuration at startup, not runtime
4. **Provide clear defaults** - Most config should be optional
5. **Document all options** - Include examples in .env.example
6. **Use structured validation** - Return all errors at once, not one by one
7. **Environment-specific behavior** - Adjust logging, security based on environment

## Testing Configuration

```go
// internal/config/config_test.go
func TestLoadConfig(t *testing.T) {
    // Set test environment variables
    t.Setenv("DATABASE_URL", "file:test.db")
    t.Setenv("SESSION_KEY", "test-key-minimum-32-characters-long")
    t.Setenv("DATABASE_DRIVER", "go-libsql")

    cfg, err := Load()
    assert.NoError(t, err)
    assert.Equal(t, "file:test.db", cfg.DatabaseURL)
    assert.Equal(t, "go-libsql", cfg.DatabaseDriver)
}

func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr string
    }{
        {
            name:    "missing database URL",
            config:  Config{SessionKey: strings.Repeat("a", 32)},
            wantErr: "DATABASE_URL is required",
        },
        {
            name:    "short session key",
            config:  Config{DatabaseURL: "test.db", SessionKey: "short"},
            wantErr: "SESSION_KEY must be at least 32 characters",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.wantErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Next Steps

- Continue to [Background Jobs →](./10_background_jobs.md)
- Back to [← External Services](./8_external_services.md)
- Return to [Summary](./0_summary.md)
