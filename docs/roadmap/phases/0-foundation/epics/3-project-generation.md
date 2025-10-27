# Epic 3: Project Generation

[← Back to Phase 0](../0-foundation.md) | [← Epic 2](./2-template-engine.md) | [Epic 4 →](./4-generated-tooling.md)

## Overview

Implement the `tracks new` command to generate complete project structures. This is the core value proposition of Phase 0 - developers can create a new Go web application with a single command. This epic brings together the CLI infrastructure and template system to create functional projects.

## Goals

- Working `tracks new` command that generates complete, runnable projects
- Database driver selection during project creation
- Clean directory structure matching PRD specifications
- Valid Go module initialization
- Basic server with health check endpoint
- Generated tests that pass immediately
- Optional git repository initialization
- Robust error handling for edge cases
- Integration tests that verify generated projects actually work

## Scope

### In Scope

- `tracks new <project>` command implementation
- Database driver flag (`--db go-libsql|sqlite3|postgres`)
- Module name flag (`--module github.com/user/app`)
- Git initialization flag (`--no-git`)
- Complete directory structure creation
- Configuration file generation (go.mod, tracks.yaml, .env.example)
- Basic server implementation with health check endpoint
- Test file generation (main_test.go, health_test.go)
- Validation (project name, module name, existing directories)
- User-friendly error messages and prompts
- Integration tests that verify generated code runs successfully

### Out of Scope

- Interactive prompts (TUI mode) - defer to later phase
- Custom template selection - use defaults only
- Database schema generation - that's Phase 2
- Full application code - just scaffold structure with health endpoint
- Migration files - Phase 2
- Advanced test utilities - Phase 5 (we generate basic tests only)

## Task Breakdown

The following tasks will become GitHub issues, organized by phase:

### Phase 1: Interfaces & Types

1. **Define ProjectGenerator interface** (#90)
2. **Define ProjectConfig struct with all options (project name, module, db driver, git)** (#91)
3. **Define Validator interface for validation logic** (#92)
4. **Define error types for generation failures** (#93)
5. **Write unit tests for type definitions** (#94)

### Phase 2: Validation Logic

1. **Implement project name validation (lowercase, alphanumeric, hyphens/underscores, no spaces)** (#95)
2. **Write unit tests for project name validation edge cases** (#96)
3. **Implement module name validation (valid Go import path)** (#97)
4. **Write unit tests for module name validation** (#98)
5. **Implement directory existence validation** (#99)
6. **Write unit tests for directory validation (exists, empty, permissions)** (#100)
7. **Implement database driver validation (go-libsql, sqlite3, postgres)** (#101)
8. **Write unit tests for driver validation** (#102)

### Phase 3: Command Implementation

1. **Implement `tracks new` Cobra command structure** (#103)
2. **Wire up --db flag with validation** (#104)
3. **Wire up --module flag with validation** (#105)
4. **Wire up --no-git flag** (#106)
5. **Write unit tests for command flag parsing** (#107)
6. **Write unit tests for flag validation integration** (#108)

### Phase 4: Directory & File Generation

**Note:** Tasks in this phase create **template files** that the `tracks` generator uses to produce user projects. The "unit tests" here test the template rendering logic in the tracks codebase itself, not the generated code. Generated projects will include their own test files (main_test.go, health_test.go) created from these templates.

1. **Create directory structure (cmd/server, internal/interfaces, internal/domain/health, internal/infra/http, internal/routes, db)** (#109)
2. **Write unit tests for directory creation** (#110)
3. **Create go.mod template** (#111)
4. **Write unit tests for go.mod** (#112)
5. **Create tracks.yaml template** (#113)
6. **Write unit tests for tracks.yaml** (#114)
7. **Create .gitignore template** (#115)
8. **Write unit tests for .gitignore** (#116)
9. **Create .env.example template** (#117)
10. **Write unit tests for .env.example** (#118)
11. **Create interfaces template (internal/interfaces/health.go.tmpl)** (#119)
12. **Write unit tests for interfaces template** (#120)
13. **Create health service template (internal/domain/health/service.go.tmpl)** (#121)
14. **Write unit tests for health service template** (#122)
15. **Create routes constants template (internal/routes/routes.go.tmpl)** (#123)
16. **Write unit tests for routes constants template** (#124)
17. **Create handler template (internal/infra/http/handlers/health.go.tmpl)** (#125)
18. **Write unit tests for handler template** (#126)
19. **Create .mockery.yaml template** (#127)
20. **Write unit tests for .mockery.yaml template** (#128)

### Phase 5: Server & Main Files

1. **Create server.go template with dependency injection pattern** (#129)
2. **Create routes.go template with marker comments** (#130)
3. **Write unit tests for server and routes templates** (#131)
4. **Create main.go template with run() pattern and markers** (#132)
5. **Write unit tests for main.go template** (#133)
6. **Create db/db.go template with connection logic** (#134)

### Phase 6: Database & Config Files

1. **Create sqlc.yaml template (output: db/generated)** (#135)
2. **Create README.md template** (#136)
3. **Write unit tests for config file templates** (#137)
4. **Create Makefile template with mocks target** (#138)

### Phase 7: Git Initialization & Output

1. **Implement git initialization logic (respecting --no-git)** (#139)
2. **Write unit tests for git init with and without flag** (#140)
3. **Implement post-generation success output with next steps** (#141)
4. **Write unit tests for output rendering** (#142)

### Phase 8: Integration & Runtime Verification

**Note:** These integration tests verify generated projects compile, run, and respond to HTTP requests without requiring database connectivity. The health check endpoint is database-free by design. Database integration testing (migrations, SQLC, queries) will be covered in Epic 4: Generated Project Tooling.

1. **Create integration test that generates full project** (#143)
2. **Integration test: verify `go mod download` succeeds** (#144)
3. **Integration test: verify `go test ./...` passes on generated project** (#145)
4. **Integration test: verify `go build ./cmd/server` succeeds** (#146)
5. **Integration test: run server binary and verify it starts** (#147)
6. **Integration test: hit health check endpoint, verify 200 OK response** (#148)
7. **Add cross-platform integration tests (Linux, macOS, Windows)** (#149)
8. **Document `tracks new` command and all flags** (#150)

## Dependencies

### Prerequisites

- Epic 1 (CLI Infrastructure) - need command framework
- Epic 2 (Template Engine) - need templates to render
- Go module system understanding

### Blocks

- Epic 4 (Tooling) - generates projects that tooling will enhance
- All future features that extend project generation

## Acceptance Criteria

### Basic Generation

- [ ] `tracks new myapp` creates valid project structure
- [ ] `tracks new myapp --db postgres` sets PostgreSQL driver
- [ ] `tracks new myapp --module github.com/me/app` uses custom module
- [ ] `tracks new myapp --no-git` skips git initialization
- [ ] Generated go.mod has correct module name and Go version
- [ ] Generated project directory matches PRD structure
- [ ] Generated tracks.yaml has sensible defaults

### Validation

- [ ] Command fails gracefully if directory already exists
- [ ] Command validates project name (no spaces, special chars)
- [ ] Command validates module name (valid Go import path)

### Generated Code Quality

- [ ] Generated project has test files (main_test.go, health_test.go)
- [ ] Generated project's tests pass (`go test ./...` succeeds)
- [ ] Generated server binary builds successfully (`go build ./cmd/server`)
- [ ] Generated server runs successfully (binary executes without errors)
- [ ] Health check endpoint responds correctly (GET /health returns 200 OK)

### Integration Testing

- [ ] Integration test generates full project
- [ ] Integration test runs `go mod download` successfully
- [ ] Integration test runs `go test ./...` (all tests pass)
- [ ] Integration test builds server binary
- [ ] Integration test runs server and hits health endpoint
- [ ] Integration test verifies graceful shutdown
- [ ] Works on Linux, macOS, and Windows

### User Experience

- [ ] Post-generation shows helpful next steps
- [ ] Error messages are clear and actionable

## Technical Notes

### Command Structure

```go
var newCmd = &cobra.Command{
    Use:   "new [project]",
    Short: "Create a new Tracks application",
    Args:  cobra.ExactArgs(1),
    RunE:  runNew,
}

func init() {
    newCmd.Flags().String("db", "go-libsql", "Database driver (go-libsql, sqlite3, postgres)")
    newCmd.Flags().String("module", "", "Custom module name (default: project name)")
    newCmd.Flags().Bool("no-git", false, "Skip git initialization")
}
```

### Directory Structure to Generate

```text
myapp/
├── cmd/server/
│   ├── main.go              # run() pattern with markers
│   └── main_test.go         # Tests run() with random ports
├── internal/
│   ├── interfaces/
│   │   └── health.go        # HealthService interface
│   ├── domain/health/
│   │   └── service.go       # Implements interfaces.HealthService
│   ├── infra/http/
│   │   ├── server.go        # Server struct with DI
│   │   ├── routes.go        # Route registration with markers
│   │   └── handlers/
│   │       └── health.go    # Handler methods on server
│   └── routes/
│       └── routes.go        # const APIHealth = "/api/health"
├── db/
│   ├── db.go                # Connection logic
│   ├── migrations/          # Empty (for future)
│   ├── queries/             # Empty (for future)
│   └── generated/           # Empty (for future SQLC)
├── test/mocks/              # Empty (mockery generates here)
├── go.mod
├── tracks.yaml
├── .mockery.yaml            # Mockery configuration
├── sqlc.yaml                # SQLC configuration
├── .gitignore
├── .env.example
├── Makefile                 # Includes 'make mocks' target
└── README.md
```

**Key Points:**

- Minimal but architecturally correct structure
- Supports incremental generation via markers
- No import cycles (interfaces package isolated)
- Database package ready for SQLC/Goose
- Mockery configured to auto-discover interfaces

### Security: Environment Variables

Generated projects must handle secrets securely:

- `.gitignore` must include `.env` to prevent committing secrets
- `.env.example` should be generated with placeholder values and a warning comment:

```bash
# WARNING: Never commit .env file with real secrets!
# Copy this file to .env and fill in your actual values.
DATABASE_URL=postgres://user:password@localhost:5432/dbname
SECRET_KEY=your-secret-key-here
```

The generated `.gitignore` should include:

```gitignore
# Environment files with secrets
.env
.env.local
.env.*.local
```

### Validation Rules

- **Project name:** lowercase, alphanumeric, hyphens/underscores allowed, no spaces
- **Module name:** valid Go import path format
- **Directory:** must not exist, or be empty

### Error Handling

Provide clear, actionable error messages with proper error wrapping:

```go
// Example error handling pattern
if err := os.MkdirAll(projectDir, 0755); err != nil {
    return fmt.Errorf("failed to create project directory: %w", err)
}

if !isValidProjectName(name) {
    return fmt.Errorf("invalid project name '%s': use lowercase letters, numbers, hyphens, and underscores", name)
}

if _, err := os.Stat(projectDir); err == nil {
    return fmt.Errorf("directory '%s' already exists: use a different name or remove the directory", projectDir)
}
```

Always use `fmt.Errorf` with `%w` to wrap errors, providing context while preserving the error chain for debugging.

### Cross-Platform Path Handling

Use `filepath` package functions for cross-platform compatibility:

```go
// Use filepath.Join for building paths
projectDir := filepath.Join(baseDir, projectName)
configFile := filepath.Join(projectDir, "tracks.yaml")

// Use filepath.FromSlash for template paths
templatePath := filepath.FromSlash("internal/templates/project")
```

Never construct paths with string concatenation or hardcoded slashes. The `filepath` package handles platform-specific separators (\ on Windows, / on Unix).

## Architecture Patterns

### Package Structure

```text
internal/
├── interfaces/         # All interfaces (zero dependencies)
│   ├── health.go      # HealthService interface
│   └── common.go      # Shared interfaces
├── domain/            # Business logic by feature
│   └── health/
│       └── service.go # Implements interfaces.HealthService
├── infra/http/
│   ├── server.go      # Server struct
│   ├── routes.go      # Route registration with markers
│   └── handlers/
│       └── health.go  # Handler methods
└── routes/
    └── routes.go      # Route constants

db/
├── db.go              # Connection logic
├── migrations/        # Goose migrations
├── queries/           # SQL source files
└── generated/         # SQLC output
```

### Interface-First Pattern

All interfaces defined in `internal/interfaces/` to prevent import cycles with mockery.

```go
// internal/interfaces/health.go
package interfaces

import (
    "context"
    "time"
)

type HealthService interface {
    Check(ctx context.Context) HealthStatus
}

type HealthStatus struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
}
```

### Main.go Pattern (Mat Ryer)

```go
// cmd/server/main.go
func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    // TRACKS:DB:BEGIN
    database, err := db.New(cfg.DatabaseURL)
    if err != nil {
        return fmt.Errorf("connect db: %w", err)
    }
    defer database.Close()
    // TRACKS:DB:END

    // TRACKS:SERVICES:BEGIN
    healthService := health.NewService()
    // TRACKS:SERVICES:END

    srv := http.NewServer(cfg).
        WithHealthService(healthService).
        RegisterRoutes()

    return srv.ListenAndServe()
}
```

### Incremental Generation Markers

**Purpose**: Allow `tracks generate resource` to safely update existing files.

**Marker Pattern**:

```go
// TRACKS:SECTION_NAME:BEGIN
// Generated code goes here
// TRACKS:SECTION_NAME:END
```

**Sections in main.go**:

- `DB` - Database connection
- `REPOSITORIES` - Repository instantiation
- `SERVICES` - Service instantiation

**Sections in routes.go**:

- `API_ROUTES` - /api/* routes (JSON only)
- `WEB_ROUTES` - Public HTML routes
- `PROTECTED_ROUTES` - Auth-required routes

### Database Package

```go
// db/db.go
package db

import "database/sql"

func New(dsn string) (*sql.DB, error) {
    db, err := sql.Open("libsql", dsn)
    if err != nil {
        return nil, err
    }
    return db, db.Ping()
}
```

**SQLC Configuration** (sqlc.yaml):

```yaml
version: "2"
sql:
  - schema: "db/migrations"
    queries: "db/queries"
    engine: "sqlite"
    gen:
      go:
        package: "generated"
        out: "db/generated"
```

**Mockery Configuration** (.mockery.yaml):

```yaml
with-expecter: true
dir: "internal/interfaces"
output: "test/mocks/{{.InterfaceName}}.go"
outpkg: mocks
all: true  # Auto-discover all interfaces
```

### Integration Test Requirements

Integration tests must verify the complete project lifecycle:

1. **Generate**: Run `tracks new testapp` successfully
2. **Dependencies**: `go mod download` succeeds
3. **Tests Pass**: `go test ./...` exits 0
4. **Build**: `go build ./cmd/server` produces binary
5. **Run**: Binary starts without errors
6. **Health Check**: `curl http://localhost:8080/health` returns 200 OK
7. **Shutdown**: Server stops gracefully

This ensures every generated project is immediately usable and confidence-inspiring.

## Testing Strategy

Following Epic 1 and Epic 2's incremental testing approach:

**Pattern:** Interface Definition → Implementation → Unit Test → Integration Test

Each component follows this cycle:

```text
1. Define Interface (Phase 1)
   ↓
2. Implement Component (Phase 2-7)
   ↓
3. Write Unit Tests (immediately after implementation)
   ↓
4. Integration Tests (Phase 8 - verify all pieces work together)
```

### Unit Tests (Phases 1-7)

- Test type definitions and interfaces
- Test validation logic (project name, module name, directory, driver)
- Test command flag parsing
- Test file generation with various configurations
- Test template rendering for all generated files
- Test git initialization with and without --no-git flag
- Test output rendering

### Integration Tests (Phase 8)

- **Full Project Generation**: Generate complete project with `tracks new`
- **Dependency Installation**: Verify `go mod download` succeeds
- **Generated Tests Pass**: Run `go test ./...`, verify exit code 0
- **Build Success**: Run `go build ./cmd/server`, verify binary created
- **Server Runtime**: Start binary, verify it runs without errors
- **Health Check**: Hit `/health` endpoint, verify 200 OK with correct JSON
- **Graceful Shutdown**: Verify server stops cleanly
- **Cross-Platform**: Test on Linux, macOS, Windows
- **Flag Combinations**: Test all flag variations (--db, --module, --no-git)
- **Error Cases**: Test invalid names, existing directories, invalid drivers

## Next Epic

[Epic 4: Generated Project Tooling →](./4-generated-tooling.md)
