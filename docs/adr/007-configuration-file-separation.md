# ADR 007: Configuration File Separation

**Status:** Accepted
**Date:** 2025-11-03
**Deciders:** Engineering Team
**Related Issues:** #134, #113, #114, #219, #137, #143, #150
**Related ADRs:** None

## Context

During implementation of database connection configuration (#134), we discovered significant confusion throughout the codebase regarding the purpose of `tracks.yaml`. The file was being used for **both**:

1. Tracks CLI project metadata (database driver, module path, project name)
2. Generated application runtime configuration (server settings, database URL, session config)

This conflation creates several problems:

### Problem 1: Security Risk

Runtime configuration often contains secrets (database URLs, session keys, API credentials). If stored in a committed YAML file, these secrets could be accidentally committed to version control.

### Problem 2: Separation of Concerns

The Tracks CLI needs different information than the running application:

- **CLI needs**: Database driver (for code generation), module path (for imports), resource registry (for incremental generation)
- **Application needs**: Server port, timeouts, database connection string, logging level

### Problem 3: Developer Confusion

Developers don't know where to put settings. Is `tracks.yaml` for the CLI or the app? Both? Where do secrets go?

### Problem 4: Version Tracking

The CLI needs to track which version of Tracks created/last upgraded the project for migration purposes. Runtime config doesn't need this.

### Problem 5: Inconsistent Documentation

- CLAUDE.md line 210: Implies `tracks.yaml` is runtime config
- PRD 9: Shows config loader reading `tracks.yaml` for runtime
- PRD 9: Also shows `.env` for runtime (conflicting)
- Epic 3 tasks: Reinforce wrong pattern

## Decision

We will **separate configuration into two distinct files** with a clear principle:

### Core Principle: Configuration for Generation, Not State Reflection

**`.tracks.yaml` contains settings for HOW to build, not WHAT was built.**

- ✅ Store: Tooling preferences, generation configuration, developer workflow settings
- ❌ Don't store: Lists of generated resources, insertion point locations, file manifests

**Why?** Code is the source of truth. When developers edit generated code (which they should!), a resource registry in `.tracks.yaml` becomes stale immediately. Instead, we use AST parsing and comment markers to discover what exists.

### 1. `.tracks.yaml` (dotfile) - Tracks CLI Project Metadata & Tooling Configuration

**Purpose:** Machine-readable project metadata and generation/tooling preferences
**Read by:** `tracks` CLI commands (`generate`, `db migrate`, `upgrade`, TUI, etc.)
**Written by:** `tracks new` (initial), `tracks config` (updates), developers (manual)
**Committed to Git:** ✅ Yes
**Contains secrets:** ❌ No

### 2. `.env` - Application Runtime Configuration

**Purpose:** Environment-specific runtime configuration for generated applications
**Read by:** Generated application at startup (via Viper)
**Written by:** Developers (manually or via tooling)
**Committed to Git:** ❌ No (`.env` is gitignored, `.env.example` is committed)
**Contains secrets:** ✅ Yes (database URLs, session keys, API credentials)

## `.tracks.yaml` Schema

### Phase 0 (v0.1.0-v0.2.0)

```yaml
# Schema version for .tracks.yaml format migrations
schema_version: "1.0"

# Project metadata (immutable after creation)
project:
  name: "myapp"
  module_path: "github.com/user/myapp"
  created_at: "2025-11-03T22:15:30Z"
  tracks_version: "v0.1.0"           # Version that created this project
  last_upgraded_version: "v0.1.0"    # Last version that ran migrations
  database_driver: "go-libsql"        # Immutable: go-libsql, sqlite3, postgres
```

**Commands that interact:**

- `tracks new` - **WRITES** all fields at project creation
- `tracks version` - **READS** `tracks_version` to check compatibility
- All future commands - **READ** `database_driver` to generate correct SQL

### Phase 4 (TUI & Generation - v0.5.0+)

```yaml
schema_version: "1.0"

project:
  # ... (same as Phase 0)

# TUI preferences (mix console configuration)
tui:
  theme: "dark"                    # UI color scheme
  default_view: "dashboard"        # Start view: dashboard, logs, db, jobs
  log_filter_level: "info"         # Default log level filter
  refresh_interval: "1s"           # Live update frequency

# Code generation preferences
generation:
  template_version: "v0.5.0"
  custom_templates:
    service: "~/.tracks/templates/service.go.tmpl"     # User-provided template
    repository: "~/.tracks/templates/repository.go.tmpl"
    handler: ""                                         # Empty = use built-in

  conventions:
    naming_style: "snake_case"     # File naming: snake_case, kebab-case, camel_case
    test_suffix: "_test"           # Test file suffix
    mock_prefix: "Mock"            # Mock struct prefix

  defaults:
    include_opentelemetry: true    # Always add tracing
    use_repository_pattern: true   # Generate repository layer
    generate_tests: true           # Auto-generate test files

# Developer workflow preferences
dev:
  hot_reload: true                 # Use Air for hot reload
  default_db: "development"        # Default database for commands
  auto_migrate: false              # Run migrations automatically
```

**Commands that interact:**

- `tracks config set tui.theme light` - **WRITES** TUI preferences
- `tracks generate` (in TUI) - **READS** generation defaults
- `tracks` (no args) - **READS** TUI preferences for dashboard layout

### Phase 2 (Data Layer - v0.3.0)

```yaml
# Migration tracking
database:
  migrations_path: "internal/db/migrations"
  last_migration: "20250103_143000_create_users"

# Tool versions for regeneration
tools:
  sqlc_version: "1.25.0"
  goose_version: "3.18.0"
```

**Commands that interact:**

- `tracks db migrate` - **WRITES** `last_migration` after each migration
- `tracks db status` - **READS** `last_migration` to show current state
- `tracks generate sql` - **READS** `sqlc_version` for compatibility

### Future Phases (v0.6.0+)

```yaml
# Feature flags for enabled capabilities
features:
  auth: true
  jobs: false
  i18n: false
  admin_panel: true

# MCP server configuration
mcp:
  enabled: true
  port: 3000

# Extension system
extensions:
  enabled:
    - "tracks-graphql"             # Enable GraphQL extension
    - "tracks-websockets"          # Enable WebSocket support

  # Extension-specific configuration
  graphql:
    schema_path: "schema.graphql"
    playground: true

# Custom middleware configuration
middleware:
  custom_chains:
    api: ["cors", "auth", "ratelimit", "logging"]
    web: ["csrf", "session", "logging"]
```

## `.env` Schema

### Phase 0 (v0.1.0-v0.2.0)

```bash
# ============================================================================
# Application Runtime Configuration
# Copy to .env for development: cp .env.example .env
# The .env file is gitignored to prevent accidentally committing secrets
# ============================================================================

# Environment
APP_ENVIRONMENT=development

# Server Configuration
APP_SERVER_PORT=:8080
APP_SERVER_READ_TIMEOUT=15s
APP_SERVER_WRITE_TIMEOUT=15s
APP_SERVER_IDLE_TIMEOUT=60s
APP_SERVER_SHUTDOWN_TIMEOUT=30s

# Database Configuration
DATABASE_URL=file:./myapp.db
APP_DATABASE_CONNECT_TIMEOUT=10s
APP_DATABASE_MAX_OPEN_CONNS=25
APP_DATABASE_MAX_IDLE_CONNS=5
APP_DATABASE_CONN_MAX_LIFETIME=5m

# Logging Configuration
APP_LOGGING_LEVEL=info
APP_LOGGING_FORMAT=json

# Session Configuration
APP_SESSION_LIFETIME=24h
APP_SESSION_COOKIE_NAME=session_id
APP_SESSION_COOKIE_SECURE=false
APP_SESSION_COOKIE_HTTP_ONLY=true
APP_SESSION_COOKIE_SAME_SITE=lax

# Secrets (NEVER COMMIT THESE VALUES)
SECRET_KEY=your-secret-key-here-replace-with-secure-random-value
```

### Future Phases

Additional sections to be added as features are implemented:

- Email provider configuration (Phase 3)
- SMS provider configuration (Phase 3)
- Storage provider configuration (Phase 3)
- Queue configuration (Phase 4)
- Rate limiting (Phase 5)
- OAuth credentials (Phase 3)
- Observability endpoints (Phase 6)

## Code Discovery Strategy

### Why NOT Store Generated Resources in `.tracks.yaml`

**Problem with resource registries:** They become stale the moment a developer edits generated code (which they should!).

**Example of the problem:**

```yaml
# .tracks.yaml says this exists:
resources:
  - name: "post"
    handlers: ["Create", "Show", "Update", "Delete"]

# But developer deleted DeletePost handler and added CustomAction handler
# .tracks.yaml is now wrong - creates drift and confusion
```

### Our Approach: Code is the Source of Truth

We use **two complementary strategies** to discover what exists in the codebase:

#### 1. Comment Markers for Insertion Points

Generated code includes special comments that mark where to insert new code:

```go
// cmd/server/main.go
func run() error {
    // TRACKS:DB:BEGIN
    database, err := db.New(ctx, cfg.Database)
    // TRACKS:DB:END

    // TRACKS:REPOSITORIES:BEGIN
    // Generated repositories will be inserted here
    // TRACKS:REPOSITORIES:END

    // TRACKS:SERVICES:BEGIN
    healthService := health.NewService()
    // TRACKS:SERVICES:END
}
```

**Benefits:**

- Self-documenting in the code itself
- Can't get out of sync with the file they're in
- Works regardless of what's between the markers
- Developers can customize the marker names if desired (stored in `.tracks.yaml`)

#### 2. AST Parsing for Resource Discovery

When we need to know what resources exist, we parse the actual Go code:

```go
// tracks generate service user --depends-on post
// Step 1: Parse internal/domain/*/service.go to find existing services
// Step 2: Analyze imports and dependencies
// Step 3: Generate new code that integrates correctly
```

**What we discover via AST parsing:**

- Existing domain services and their interfaces
- Repository implementations
- Handler functions and their routes
- DTO structs and validation tags
- Database models and relationships

**Benefits:**

- Always accurate (parses actual code)
- Handles manual edits gracefully
- Enables intelligent code generation
- No drift between config and reality

### Hybrid Approach for Performance

For frequently-used data (like custom template paths or generation preferences), store in `.tracks.yaml`. For code structure discovery, always parse the actual files.

## Configuration Loading Pattern

### Generated Application Config Loader

```go
func Load() (*Config, error) {
    v := viper.New()

    // Set defaults (fallback values)
    v.SetDefault("server.port", ":8080")
    v.SetDefault("server.read_timeout", "15s")
    v.SetDefault("server.write_timeout", "15s")
    v.SetDefault("server.idle_timeout", "60s")
    v.SetDefault("server.shutdown_timeout", "30s")
    v.SetDefault("logging.level", "info")
    v.SetDefault("logging.format", "json")
    v.SetDefault("database.connect_timeout", "10s")
    v.SetDefault("database.max_open_conns", 25)
    v.SetDefault("database.max_idle_conns", 5)
    v.SetDefault("database.conn_max_lifetime", "5m")

    // Read from .env file (optional, for development)
    v.SetConfigFile(".env")
    _ = v.ReadInConfig()  // Ignore error if file doesn't exist

    // Environment variables override everything
    v.SetEnvPrefix("APP")
    v.AutomaticEnv()

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return &cfg, nil
}
```

**Key points:**

- ❌ Does NOT read `.tracks.yaml`
- ✅ Reads `.env` file if present (development)
- ✅ Environment variables override everything (production)
- ✅ Falls back to sensible defaults

### Hierarchical Configuration (Lowest to Highest Priority)

1. **Default values** (in code via `v.SetDefault()`)
2. **`.env` file** (development only, gitignored)
3. **Environment variables** (production, prefixed with `APP_`)

Example:

```bash
# Database URL resolution:
# 1. Default: Not set (would error)
# 2. .env file: DATABASE_URL=file:./local.db
# 3. Environment: export DATABASE_URL=libsql://prod.turso.io
# Final value: libsql://prod.turso.io (env var wins)
```

## File Purposes Table

| File | Purpose | Read By | Contains | Committed to Git | Has Secrets |
|------|---------|---------|----------|------------------|-------------|
| `.tracks.yaml` | Tracks CLI metadata & tooling prefs | `tracks` CLI | Driver, module, versions, TUI prefs, gen config | ✅ Yes | ❌ No |
| `.env` | Development runtime config | Generated app | DB URLs, ports, timeouts, keys | ❌ No | ✅ Yes |
| `.env.example` | Runtime config template | Developers | Placeholder values, documentation | ✅ Yes | ❌ No |
| Environment variables | Production runtime config | Generated app (via Viper) | All runtime settings | N/A | ✅ Yes |
| Source code | What was actually generated | AST parser | Resources, handlers, services, repos | ✅ Yes | ❌ No |

## Version Tracking and Migration Strategy

### Version Fields in `.tracks.yaml`

- **`tracks_version`**: Immutable, set at project creation, identifies original Tracks version
- **`last_upgraded_version`**: Mutable, updated by `tracks upgrade`, tracks last successful migration
- **`schema_version`**: Format version of `.tracks.yaml` itself (for breaking schema changes)

### Migration Use Cases

#### Scenario 1: User upgrades Tracks CLI

```bash
$ tracks upgrade --check
🔍 Detected project created with v0.1.0 (current CLI: v0.3.0)
📋 Available migrations:
  • v0.1.0 → v0.2.0: Add asset pipeline configuration
  • v0.2.0 → v0.3.0: Add SQLC/Goose integration

$ tracks upgrade
✅ Migrated v0.1.0 → v0.2.0
✅ Migrated v0.2.0 → v0.3.0
✅ Updated .tracks.yaml to v0.3.0
```

#### Scenario 2: Schema version changes

```bash
$ tracks upgrade
🔄 Upgrading .tracks.yaml schema 1.0 → 2.0
  • Restructuring database config
  • Adding connection pool settings
✅ Schema upgraded to 2.0
```

### Compatibility Rules

- New CLI versions CAN work with old projects (with warnings to upgrade)
- Old CLI versions CANNOT work with new projects (error with clear message)
- `tracks generate` checks compatibility before modifying code
- Team members should upgrade Tracks CLI together to avoid conflicts

## Consequences

### Positive

1. **Clear separation of concerns**: CLI metadata vs runtime config
2. **Security**: Secrets stay in `.env` (gitignored), never committed
3. **Follows 12-factor app principles**: Configuration via environment
4. **Better DX**: Developers know where to look for each type of setting
5. **Migration support**: Version tracking enables safe upgrades
6. **No drift**: AST parsing ensures code is source of truth, not config files
7. **Dotfile convention**: `.tracks.yaml` clearly indicates tooling metadata
8. **Customization**: TUI preferences and custom templates support advanced workflows
9. **DAW metaphor**: Mix console preferences align with project vision

### Negative

1. **Breaking change**: Existing templates/docs/tests need updating
2. **Migration needed**: Projects already generated with wrong pattern
3. **Two files to manage**: `.tracks.yaml` + `.env` (but with clear purposes)
4. **Learning curve**: Users must understand the distinction

### Neutral

1. **File count**: Two config files instead of one (but clearer)
2. **Documentation burden**: Must document both files clearly

## Implementation Plan

### Phase 1: Foundation (Immediate)

1. **Create this ADR** ✅
2. **Rename template**: `tracks.yaml.tmpl` → `.tracks.yaml.tmpl`
3. **Rewrite `.tracks.yaml.tmpl`**: CLI metadata only (Phase 0 schema)
4. **Update `config.go.tmpl`**: Remove `.tracks.yaml` reading
5. **Expand `.env.example.tmpl`**: Add all runtime settings (Phase 0 schema)
6. **Update tests**: Validate new separation
7. **Update docs**: CLAUDE.md, PRDs, roadmap

### Phase 2: Database Context Feature (#134)

Now that configuration is clarified, implement database connection pool:

1. Add `DatabaseConfig` fields in `config.go.tmpl` (pool settings)
2. Update `db.go.tmpl` (add `context.Context` parameter, pool configuration)
3. Update `main.go.tmpl` (pass context and config to `db.New()`)
4. Update tests

### Phase 3: Future Enhancements

- Add `tracks upgrade` command (migration system)
- Add `.tracks.yaml` schema versioning
- Add resource registry fields (Phase 4)
- Add migration tracking fields (Phase 2)

## References

- [12-Factor App: Config](https://12factor.net/config)
- [Viper Configuration Library](https://github.com/spf13/viper)
- Related PRD: `docs/prd/9_configuration.md`
- Related Epic: `docs/roadmap/phases/0-foundation/epics/3-project-generation.md`
- GitHub Issues: #134 (DB config), #113 (tracks.yaml), #219 (config package)

## Decision Makers

This ADR establishes the standard for configuration file separation across the Tracks project.

---

**Last Updated:** 2025-11-03
**Next Review:** After Phase 4 (Generation) implementation
