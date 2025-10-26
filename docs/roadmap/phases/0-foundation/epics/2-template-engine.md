# Epic 2: Template Engine & Embedding

[← Back to Phase 0](../0-foundation.md) | [← Epic 1](./1-cli-infrastructure.md) | [Epic 3 →](./3-project-generation.md)

## Overview

Build the template system that will power project generation. This includes setting up Go's embed system for template files, creating the initial template structure, and implementing the rendering engine with variable substitution. This epic enables Epic 3 to actually generate projects.

## Goals

- Functional template embedding using Go's embed package
- Template rendering engine with variable substitution
- Minimal but complete template set for generated projects
- Template validation and error handling
- Extensible template structure for future additions

## Scope

### In Scope

- Go embed.go setup for template files
- Template directory structure in tracks repository
- Template rendering logic with variable substitution
- Initial template set (go.mod, main.go, basic structure)
- Template validation (syntax, required variables)
- Template unit tests

### Out of Scope

- Complete template set for all features - just minimal viable templates
- Advanced templating (conditionals, loops) - keep simple
- Template hot-reload for development - static embedding is fine
- User-defined custom templates - future enhancement

## Task Breakdown

The following tasks will become GitHub issues, ordered by dependency:

### Phase 1: Define Interfaces & Types

1. **Define TemplateRenderer interface and core rendering types** [#54](https://github.com/anomalousventures/tracks/issues/54)
2. **Define TemplateData struct with variable schema** [#55](https://github.com/anomalousventures/tracks/issues/55)
3. **Define template error types and validation interfaces** [#56](https://github.com/anomalousventures/tracks/issues/56)

### Phase 2: Embed System Setup

1. **Create template directory structure in internal/templates/** [#57](https://github.com/anomalousventures/tracks/issues/57)
2. **Set up embed.go with embed directive for template files** [#58](https://github.com/anomalousventures/tracks/issues/58)
3. **Write unit tests for embedded FS file reading** [#59](https://github.com/anomalousventures/tracks/issues/59)
4. **Verify embed system works across platforms (Windows/Unix paths)** [#60](https://github.com/anomalousventures/tracks/issues/60)

### Phase 3: Basic Rendering Engine

1. **Implement basic template file reader from embed.FS** [#61](https://github.com/anomalousventures/tracks/issues/61)
2. **Implement simple variable substitution using text/template** [#62](https://github.com/anomalousventures/tracks/issues/62)
3. **Write unit tests for variable substitution with various data types** [#63](https://github.com/anomalousventures/tracks/issues/63)
4. **Add cross-platform path handling (filepath.Join, FromSlash)** [#64](https://github.com/anomalousventures/tracks/issues/64)
5. **Write unit tests for path normalization on Windows/Unix** [#65](https://github.com/anomalousventures/tracks/issues/65)

### Phase 4: Template Creation with Tests

1. **Create go.mod.tmpl template** [#66](https://github.com/anomalousventures/tracks/issues/66)
2. **Write unit tests for go.mod template rendering** [#67](https://github.com/anomalousventures/tracks/issues/67)
3. **Create .gitignore.tmpl template** [#68](https://github.com/anomalousventures/tracks/issues/68)
4. **Write unit tests for .gitignore rendering** [#69](https://github.com/anomalousventures/tracks/issues/69)
5. **Create cmd/server/main.go.tmpl template** [#70](https://github.com/anomalousventures/tracks/issues/70)
6. **Write unit tests for cmd/server/main.go rendering with different module names** [#71](https://github.com/anomalousventures/tracks/issues/71)
7. **Create tracks.yaml.tmpl configuration template** [#72](https://github.com/anomalousventures/tracks/issues/72)
8. **Write unit tests for tracks.yaml with different DB drivers** [#73](https://github.com/anomalousventures/tracks/issues/73)
9. **Create .env.example.tmpl template** [#74](https://github.com/anomalousventures/tracks/issues/74)
10. **Write unit tests for .env.example rendering** [#75](https://github.com/anomalousventures/tracks/issues/75)
11. **Create README.md.tmpl template** [#76](https://github.com/anomalousventures/tracks/issues/76)
12. **Write unit tests for README.md rendering** [#77](https://github.com/anomalousventures/tracks/issues/77)

### Phase 5: Validation & Integration

1. **Implement template validation (missing variables, syntax errors)** [#78](https://github.com/anomalousventures/tracks/issues/78)
2. **Write unit tests for validation error handling** [#79](https://github.com/anomalousventures/tracks/issues/79)
3. **Create integration test framework that renders all templates together** [#80](https://github.com/anomalousventures/tracks/issues/80)
4. **Add cross-platform integration tests (verify on Linux, macOS, Windows paths)** [#81](https://github.com/anomalousventures/tracks/issues/81)

### Phase 6: Documentation

1. **Document template system architecture, variable schema, and extension guide** [#82](https://github.com/anomalousventures/tracks/issues/82)

## Dependencies

### Prerequisites

- Epic 1 (CLI Infrastructure) - need CLI to test template rendering
- Go 1.25+ with embed support

### Blocks

- Epic 3 (Project Generation) - can't generate without templates
- Epic 4 (Tooling) - tooling templates depend on this system

## Acceptance Criteria

### Phase 1: Interfaces & Types

- [ ] TemplateRenderer interface defined with clear method signatures
- [ ] TemplateData struct includes all required variables
- [ ] Error types defined for validation and rendering failures

### Phase 2: Embed System

- [ ] Template files successfully embedded in tracks binary
- [ ] Embedded FS accessible and testable
- [ ] Works correctly on Windows and Unix systems

### Phase 3: Rendering Engine

- [ ] Can render templates with variable substitution
- [ ] Handles all TemplateData field types correctly
- [ ] Cross-platform path handling works (filepath.Join, FromSlash)
- [ ] All rendering tests pass

### Phase 4: Templates

- [ ] go.mod template generates valid Go module file
- [ ] cmd/server/main.go template creates runnable application entry point
- [ ] tracks.yaml generates valid configuration
- [ ] .gitignore includes appropriate patterns
- [ ] .env.example includes security warnings and placeholder values
- [ ] README.md provides project overview and instructions
- [ ] Each template has passing unit tests

### Phase 5: Validation & Integration

- [ ] Template rendering handles missing variables gracefully
- [ ] Validation catches syntax errors before generation
- [ ] Integration tests verify all templates work together
- [ ] Cross-platform integration tests pass

### Phase 6: Documentation

- [ ] Documentation explains template structure and variables
- [ ] Variable schema documented with examples
- [ ] Extension guide for adding new templates

## Technical Notes

### Interface-First Approach

Following Epic 1's Renderer pattern, define interfaces before implementation:

```go
// internal/generator/template/renderer.go
package template

import (
    "embed"
    "io/fs"
)

// Renderer handles template rendering with variable substitution
type Renderer interface {
    // Render renders a template by name with the given data
    Render(name string, data TemplateData) (string, error)

    // RenderToFile renders a template and writes it to a file
    RenderToFile(templateName string, data TemplateData, outputPath string) error

    // Validate checks if a template exists and is valid
    Validate(name string) error
}

// TemplateData contains all variables available to templates
type TemplateData struct {
    ModuleName  string // e.g., github.com/user/myapp
    ProjectName string // e.g., myapp
    DBDriver    string // go-libsql, sqlite3, postgres
    GoVersion   string // e.g., 1.25
    Year        int    // for copyright
}

// TemplateRenderer implements Renderer using Go's embed.FS
type TemplateRenderer struct {
    fs embed.FS
}

func NewRenderer(fs embed.FS) Renderer {
    return &TemplateRenderer{fs: fs}
}
```

### Embed Structure

```go
// internal/templates/embed.go
package templates

import "embed"

//go:embed project/**/*.tmpl
var FS embed.FS
```

### Template Variables

Common variables all templates will support:

- `{{.ModuleName}}` - Go module name (e.g., github.com/user/myapp)
- `{{.ProjectName}}` - Project name (e.g., myapp)
- `{{.DBDriver}}` - Database driver (go-libsql, sqlite3, postgres)
- `{{.GoVersion}}` - Go version to use
- `{{.Year}}` - Current year for copyright

### Template Directory Structure

```text
internal/templates/
├── project/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go.tmpl
│   ├── go.mod.tmpl
│   ├── tracks.yaml.tmpl
│   ├── README.md.tmpl
│   ├── .gitignore.tmpl
│   └── .env.example.tmpl
└── embed.go
```

Note: Templates mirror the structure of generated projects. Nested paths like `cmd/server/main.go.tmpl` will be preserved in the generated output.

### Keep It Simple

Use Go's text/template for now. Don't overcomplicate with complex template logic. The goal is simple variable substitution.

### Testing Strategy

Following Epic 1's test-as-you-go approach:

#### Unit Tests (Phase 2: 3, Phase 3: 3, 5, Phase 4: 2, 4, 6, 8, 10, 12, Phase 5: 2)

- Test each component immediately after implementation
- Test template rendering with various variable combinations
- Test missing/invalid variables
- Test cross-platform path handling
- Test validation logic

#### Integration Tests (Phase 5: 3-4)

- Test that all templates render together correctly
- Verify generated files are valid (go.mod parses, etc.)
- Test on multiple platforms (Linux, macOS, Windows)
- Verify filepath handling across OS

#### Test Pattern

```go
func TestGoModTemplate(t *testing.T) {
    renderer := NewRenderer(templates.FS)
    data := TemplateData{
        ModuleName: "github.com/user/myapp",
        GoVersion:  "1.25",
    }

    result, err := renderer.Render("go.mod.tmpl", data)
    require.NoError(t, err)
    assert.Contains(t, result, "module github.com/user/myapp")
    assert.Contains(t, result, "go 1.25")
}
```

### Template File Naming Convention

All template files use the `.tmpl` extension. The template engine strips this extension when writing output files:

- `go.mod.tmpl` → `go.mod`
- `.gitignore.tmpl` → `.gitignore`
- `main.go.tmpl` → `main.go`

Dotfiles are preserved: the leading dot in `.gitignore.tmpl` is kept, resulting in `.gitignore` in the generated project.

## Testing Strategy (Summary)

Following Epic 1's incremental testing approach:

- **Unit tests after each component** (Tasks 6, 10, 12, 14, 16, 18, 20, 22)
- **Test template rendering** with various variable combinations
- **Test error handling** for missing/invalid variables
- **Test cross-platform paths** (Windows vs Unix)
- **Integration tests** that render all templates together (Tasks 23-24)
- **Validation tests** that generated files are valid (go.mod parses, etc.)

See Technical Notes → Testing Strategy for detailed testing patterns.

## Next Epic

[Epic 3: Project Generation →](./3-project-generation.md)
