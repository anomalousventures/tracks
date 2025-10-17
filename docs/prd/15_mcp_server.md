# MCP Server

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks provides a Model Context Protocol (MCP) server that exposes framework operations to AI assistants. The server enables consistent code generation, project analysis, and development workflows through AI tooling. Available as both a Docker container and standalone binary, it supports stdio and HTTP transports.

## Goals

- Expose framework operations via Model Context Protocol
- Enable AI assistants to generate and analyze code
- Docker distribution for easy integration
- Support both stdio and HTTP transports
- Comprehensive toolset for development workflows

## User Stories

- As a developer using AI, I want my assistant to generate Tracks code
- As a developer, I want AI to analyze my project structure
- As a team lead, I want consistent code generation across the team
- As a developer, I want AI to help debug and trace issues
- As a developer, I want easy MCP server setup

## MCP Server Architecture

```go
// cmd/mcp/main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/gofrs/uuid/v5"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/modelcontextprotocol/go-sdk/transport/stdio"
)

type TracksServer struct {
    projectPath string
    config      *Config
    generator   *CodeGenerator
    analyzer    *ProjectAnalyzer
}

func main() {
    server := &TracksServer{
        projectPath: os.Getenv("PROJECT_PATH"),
    }

    transport := stdio.NewTransport(os.Stdin, os.Stdout)
    runner := mcp.NewRunner(server)

    if err := runner.Run(context.Background(), transport); err != nil {
        log.Fatal(err)
    }
}

func (s *TracksServer) Initialize() error {
    // Load project configuration
    s.config = LoadConfig(s.projectPath)

    // Initialize code generator with project settings
    s.generator = NewCodeGenerator(s.config)

    // Initialize project analyzer
    s.analyzer = NewProjectAnalyzer(s.projectPath)

    return nil
}
```

## Tool Implementations

### Code Generation Tools

#### 1. Generate Handler

```go
type GenerateHandlerInput struct {
    Name     string   `json:"name"`
    Methods  []string `json:"methods"`  // GET, POST, PUT, DELETE
    AuthRequired bool `json:"auth_required"`
}

func (s *TracksServer) generateHandler(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input GenerateHandlerInput,
) (*mcp.CallToolResult, struct{ FilePath string }, error) {

    // Generate handler code
    code := s.generator.GenerateHandler(GeneratorInput{
        Name:         input.Name,
        Methods:      input.Methods,
        AuthRequired: input.AuthRequired,
    })

    // Determine file path
    fileName := toSnakeCase(input.Name) + "_handler.go"
    filePath := fmt.Sprintf("%s/internal/http/handlers/%s",
        s.projectPath, fileName)

    // Write file
    if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
        return mcp.NewErrorResult(err.Error()), struct{}{}, nil
    }

    // Generate tests
    testCode := s.generator.GenerateHandlerTest(input.Name)
    testPath := fmt.Sprintf("%s/internal/http/handlers/%s_test.go",
        s.projectPath, toSnakeCase(input.Name))
    os.WriteFile(testPath, []byte(testCode), 0644)

    return nil, struct{ FilePath string }{FilePath: filePath}, nil
}
```

#### 2. Generate Service

```go
type GenerateServiceInput struct {
    Name         string   `json:"name"`
    Methods      []Method `json:"methods"`
    Dependencies []string `json:"dependencies"`
}

type Method struct {
    Name       string `json:"name"`
    Parameters []Param `json:"parameters"`
    Returns    []string `json:"returns"`
}

func (s *TracksServer) generateService(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input GenerateServiceInput,
) (*mcp.CallToolResult, struct{ FilePath string }, error) {

    code := s.generator.GenerateService(ServiceInput{
        Name:         input.Name,
        Methods:      input.Methods,
        Dependencies: input.Dependencies,
        UseUUIDv7:    true,  // Always use UUIDv7
    })

    filePath := fmt.Sprintf("%s/internal/services/%s_service.go",
        s.projectPath, toSnakeCase(input.Name))

    if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
        return mcp.NewErrorResult(err.Error()), struct{}{}, nil
    }

    return nil, struct{ FilePath string }{FilePath: filePath}, nil
}
```

#### 3. Scaffold Resource

```go
type ScaffoldResourceInput struct {
    Name       string              `json:"name"`
    Fields     []Field             `json:"fields"`
    Relations  []Relation          `json:"relations"`
    Features   ScaffoldFeatures    `json:"features"`
}

type ScaffoldFeatures struct {
    API        bool `json:"api"`
    Web        bool `json:"web"`
    Auth       bool `json:"auth"`
    Validation bool `json:"validation"`
    Search     bool `json:"search"`
    Audit      bool `json:"audit"`
}

func (s *TracksServer) scaffoldResource(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input ScaffoldResourceInput,
) (*mcp.CallToolResult, struct{ Files []string }, error) {

    files := []string{}

    // Generate migration
    migration := s.generator.GenerateMigration(input.Name, input.Fields)
    migrationPath := s.writeMigration(migration)
    files = append(files, migrationPath)

    // Generate SQLC queries
    queries := s.generator.GenerateQueries(input.Name, input.Features.Search)
    queryPath := s.writeQueries(queries)
    files = append(files, queryPath)

    // Run SQLC to generate repository
    if err := s.runSQLCGenerate(); err != nil {
        return mcp.NewErrorResult("SQLC generation failed"), struct{}{}, nil
    }

    // Generate service
    service := s.generator.GenerateServiceFromScaffold(input)
    servicePath := s.writeService(service)
    files = append(files, servicePath)

    // Generate handlers
    if input.Features.API {
        apiHandler := s.generator.GenerateAPIHandler(input)
        apiPath := s.writeHandler("api", apiHandler)
        files = append(files, apiPath)
    }

    if input.Features.Web {
        webHandler := s.generator.GenerateWebHandler(input)
        webPath := s.writeHandler("web", webHandler)
        files = append(files, webPath)

        // Generate templ views
        views := s.generator.GenerateViews(input)
        for viewName, viewCode := range views {
            viewPath := s.writeView(viewName, viewCode)
            files = append(files, viewPath)
        }
    }

    // Update routes
    s.updateRoutes(input.Name, input.Features)

    return nil, struct{ Files []string }{Files: files}, nil
}
```

### Code Analysis Tools

#### 4. List Routes

```go
func (s *TracksServer) listRoutes(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input struct{},
) (*mcp.CallToolResult, []Route, error) {

    routes, err := s.analyzer.ExtractRoutes()
    if err != nil {
        return mcp.NewErrorResult(err.Error()), nil, nil
    }

    return nil, routes, nil
}

type Route struct {
    Method      string   `json:"method"`
    Path        string   `json:"path"`
    Handler     string   `json:"handler"`
    Middlewares []string `json:"middlewares"`
    File        string   `json:"file"`
    Line        int      `json:"line"`
}
```

#### 5. Analyze Dependencies

```go
func (s *TracksServer) analyzeDependencies(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input struct{ Package string },
) (*mcp.CallToolResult, DependencyGraph, error) {

    graph := s.analyzer.BuildDependencyGraph(input.Package)

    // Find circular dependencies
    cycles := graph.FindCycles()
    if len(cycles) > 0 {
        graph.Warnings = append(graph.Warnings,
            fmt.Sprintf("Found %d circular dependencies", len(cycles)))
    }

    return nil, graph, nil
}
```

### Database Tools

#### 6. Migration Status

```go
func (s *TracksServer) migrationStatus(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input struct{},
) (*mcp.CallToolResult, MigrationStatus, error) {

    db, err := s.getDB()
    if err != nil {
        return mcp.NewErrorResult(err.Error()), MigrationStatus{}, nil
    }

    current, err := goose.GetDBVersion(db)
    if err != nil {
        return mcp.NewErrorResult(err.Error()), MigrationStatus{}, nil
    }

    pending, err := s.getPendingMigrations(db)

    return nil, MigrationStatus{
        Current: current,
        Latest:  s.getLatestMigration(),
        Pending: pending,
    }, nil
}
```

#### 7. Generate Query

```go
func (s *TracksServer) generateQuery(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input struct{ Description string },
) (*mcp.CallToolResult, struct{ Query string; SQLC string }, error) {

    // Use AI to generate SQL from description
    sql := s.generator.GenerateSQLFromDescription(input.Description)

    // Generate SQLC annotation
    sqlc := s.generator.WrapInSQLCQuery(sql)

    return nil, struct{
        Query string
        SQLC  string
    }{
        Query: sql,
        SQLC:  sqlc,
    }, nil
}
```

### Testing Tools

#### 8. Run Tests

```go
func (s *TracksServer) runTests(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input struct{ Pattern string; Coverage bool },
) (*mcp.CallToolResult, TestResults, error) {

    args := []string{"test", "-v"}

    if input.Pattern != "" {
        args = append(args, "-run", input.Pattern)
    }

    if input.Coverage {
        args = append(args, "-cover", "-coverprofile=coverage.out")
    }

    args = append(args, "./...")

    output, err := s.runGoCommand(args...)

    results := s.parseTestOutput(output)

    if input.Coverage {
        coverage := s.parseCoverageReport()
        results.Coverage = &coverage
    }

    return nil, results, nil
}
```

### Operations Tools

#### 9. Trace Request

```go
func (s *TracksServer) traceRequest(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input struct{ RequestID string },
) (*mcp.CallToolResult, RequestTrace, error) {

    // Query OpenTelemetry for trace
    trace := s.analyzer.GetTrace(input.RequestID)

    // Build execution flow
    flow := s.analyzer.BuildExecutionFlow(trace)

    // Identify bottlenecks
    bottlenecks := s.analyzer.FindBottlenecks(trace)

    return nil, RequestTrace{
        ID:          input.RequestID,
        Spans:       trace.Spans,
        Flow:        flow,
        Bottlenecks: bottlenecks,
        Duration:    trace.Duration,
    }, nil
}
```

## Docker Distribution

```dockerfile
# Dockerfile.mcp
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o mcp-server cmd/mcp/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/mcp-server .

# MCP stdio mode by default
ENTRYPOINT ["./mcp-server"]
```

### Docker Compose Integration

```yaml
# docker-compose.mcp.yml
version: '3.8'

services:
  tracks-mcp:
    image: tracks/mcp-server:latest
    volumes:
      - ./:/project
    environment:
      - PROJECT_PATH=/project
      - MCP_MODE=http  # or stdio
      - PORT=8090
    ports:
      - "8090:8090"  # For HTTP mode
```

## Claude Desktop Integration

```json
// claude_config.json
{
  "mcpServers": {
    "tracks": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-v", "${workspaceFolder}:/project",
        "-e", "PROJECT_PATH=/project",
        "tracks/mcp-server:latest"
      ]
    }
  }
}
```

## Complete Tool List (21 Tools)

### Code Generation (6 tools)

- **generate_handler** - Create HTTP handler with tests
- **generate_service** - Create service layer with dependency injection
- **generate_repo** - Create repository with SQLC
- **generate_migration** - Create database migration
- **generate_view** - Create templ view components
- **scaffold_resource** - Full CRUD scaffolding

### Code Analysis (5 tools)

- **list_handlers** - List all HTTP handlers
- **list_services** - List all service definitions
- **list_routes** - List all registered routes
- **find_route** - Find specific route by path
- **analyze_dependencies** - Analyze import graph

### Database (4 tools)

- **migration_status** - Check migration state
- **run_migration** - Execute pending migrations
- **introspect_schema** - Inspect database schema
- **generate_query** - Generate SQL from natural language

### Testing (2 tools)

- **run_tests** - Execute test suite with options
- **check_coverage** - Generate coverage report

### Operations (4 tools)

- **start_dev_server** - Start development server
- **check_health** - Run health checks
- **view_logs** - Tail application logs
- **trace_request** - Trace request execution flow

## Security Considerations

```go
// MCP server includes security features
type SecurityConfig struct {
    // Rate limiting per tool
    RateLimits map[string]int

    // Allowed file paths for generation
    AllowedPaths []string

    // Dangerous operations require confirmation
    RequireConfirmation []string

    // Audit logging
    AuditLog bool
}

func (s *TracksServer) validatePath(path string) error {
    // Ensure path is within project
    if !strings.HasPrefix(path, s.projectPath) {
        return fmt.Errorf("path outside project: %s", path)
    }

    // Check against allowed paths
    for _, allowed := range s.config.Security.AllowedPaths {
        if strings.HasPrefix(path, allowed) {
            return nil
        }
    }

    return fmt.Errorf("path not in allowed list: %s", path)
}
```

## Usage Examples

### Generate Complete Feature

```javascript
// From Claude or other MCP client
await mcp.callTool('scaffold_resource', {
  name: 'Article',
  fields: [
    { name: 'title', type: 'string', required: true },
    { name: 'content', type: 'text', required: true },
    { name: 'author_id', type: 'uuid', required: true },
    { name: 'published_at', type: 'timestamp', required: false }
  ],
  relations: [
    { type: 'belongs_to', name: 'author', model: 'User' }
  ],
  features: {
    api: true,
    web: true,
    auth: true,
    validation: true,
    search: true,
    audit: true
  }
});
```

### Analyze Project

```javascript
// Get project overview
const routes = await mcp.callTool('list_routes', {});
const dependencies = await mcp.callTool('analyze_dependencies', {
  package: 'internal/services'
});

// Check for issues
const health = await mcp.callTool('check_health', {});
```

## Next Steps

- Continue to [TUI Mode →](./16_tui_mode.md)
- Back to [← Code Generation](./14_code_generation.md)
- Return to [Summary](./0_summary.md)
