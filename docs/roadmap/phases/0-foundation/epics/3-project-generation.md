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

1. **Define ProjectGenerator interface**
2. **Define ProjectConfig struct with all options (project name, module, db driver, git)**
3. **Define Validator interface for validation logic**
4. **Define error types for generation failures**
5. **Write unit tests for type definitions**

### Phase 2: Validation Logic

1. **Implement project name validation (lowercase, alphanumeric, hyphens/underscores, no spaces)**
2. **Write unit tests for project name validation edge cases**
3. **Implement module name validation (valid Go import path)**
4. **Write unit tests for module name validation**
5. **Implement directory existence validation**
6. **Write unit tests for directory validation (exists, empty, permissions)**
7. **Implement database driver validation (go-libsql, sqlite3, postgres)**
8. **Write unit tests for driver validation**

### Phase 3: Command Implementation

1. **Implement `tracks new` Cobra command structure**
2. **Wire up --db flag with validation**
3. **Wire up --module flag with validation**
4. **Wire up --no-git flag**
5. **Write unit tests for command flag parsing**
6. **Write unit tests for flag validation integration**

### Phase 4: Directory & File Generation

1. **Implement directory tree creation with proper structure**
2. **Write unit tests for directory creation**
3. **Implement go.mod generation using template system**
4. **Write unit tests for go.mod with different module names**
5. **Implement tracks.yaml generation**
6. **Write unit tests for tracks.yaml with different drivers**
7. **Implement .gitignore generation**
8. **Write unit tests for .gitignore**
9. **Implement .env.example generation with security warnings**
10. **Write unit tests for .env.example**
11. **Create health check handler template (internal/handlers/health.go.tmpl)**
12. **Write unit tests for health handler template**
13. **Create health check test template (internal/handlers/health_test.go.tmpl)**
14. **Write unit tests for health test template**
15. **Create server test template (cmd/server/main_test.go.tmpl)**
16. **Write unit tests for server test template**

### Phase 5: README & Main File

1. **Implement README.md generation with project name**
2. **Write unit tests for README.md**
3. **Implement cmd/server/main.go generation**
4. **Write unit tests for cmd/server/main.go**

### Phase 6: Health Check Integration

1. **Wire health check handler into main.go template**
2. **Update main.go to include health endpoint route**
3. **Write unit tests for health check integration**
4. **Update integration test to verify generated tests pass**

### Phase 7: Git Initialization & Output

1. **Implement git initialization logic (respecting --no-git)**
2. **Write unit tests for git init with and without flag**
3. **Implement post-generation success output with next steps**
4. **Write unit tests for output rendering**

### Phase 8: Integration & Runtime Verification

1. **Create integration test that generates full project**
2. **Integration test: verify `go mod download` succeeds**
3. **Integration test: verify `go test ./...` passes on generated project**
4. **Integration test: verify `go build ./cmd/server` succeeds**
5. **Integration test: run server binary and verify it starts**
6. **Integration test: hit health check endpoint, verify 200 OK response**
7. **Add cross-platform integration tests (Linux, macOS, Windows)**
8. **Document `tracks new` command and all flags**

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

Based on Core Architecture PRD:

```text
myapp/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── handlers/
│   ├── services/
│   └── middleware/
├── .env.example
├── .gitignore
├── go.mod
├── tracks.yaml
└── README.md
```

Initially keep it minimal - just the essential structure. Full structure comes with later phases.

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

### Test Generation

Generated projects must include tests to ensure they work out of the box. This builds user confidence and prevents broken scaffolds.

#### Health Check Handler

Create a simple health check endpoint that returns JSON:

```go
// internal/handlers/health.go.tmpl
package handlers

import (
    "encoding/json"
    "net/http"
)

type HealthResponse struct {
    Status string `json:"status"`
}

func Health(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}
```

#### Health Check Tests

```go
// internal/handlers/health_test.go.tmpl
package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHealth(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    w := httptest.NewRecorder()

    Health(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }

    if ct := w.Header().Get("Content-Type"); ct != "application/json" {
        t.Errorf("expected Content-Type application/json, got %s", ct)
    }
}
```

#### Server Start Tests

```go
// cmd/server/main_test.go.tmpl
package main

import (
    "context"
    "fmt"
    "net/http"
    "testing"
    "time"
)

func waitForServer(url string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        resp, err := http.Get(url)
        if err == nil {
            resp.Body.Close()
            return nil
        }
        time.Sleep(50 * time.Millisecond)
    }
    return fmt.Errorf("server not ready after %v", timeout)
}

func TestServerStarts(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    serverAddr := "127.0.0.1:8080"
    go func() {
        startServer(ctx, serverAddr)
    }()

    healthURL := fmt.Sprintf("http://%s/health", serverAddr)
    err := waitForServer(healthURL, 2*time.Second)
    if err != nil {
        t.Fatalf("server did not start: %v", err)
    }

    resp, err := http.Get(healthURL)
    if err != nil {
        t.Fatalf("health check failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected status 200, got %d", resp.StatusCode)
    }

    cancel()
    time.Sleep(100 * time.Millisecond)
}
```

**Note:** The `startServer(ctx, serverAddr)` function would be the actual server initialization code from the generated `cmd/server/main.go`, refactored to accept a context for graceful shutdown and an address for testing. For production tests, use random ports with `net.Listen("127.0.0.1:0")` to avoid port conflicts.

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
