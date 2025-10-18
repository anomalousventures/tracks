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

The following tasks will become GitHub issues:

1. **Create template directory structure in tracks repo**
2. **Set up embed.go to embed template files**
3. **Implement template rendering engine**
4. **Add variable substitution for module name, db driver, etc.**
5. **Create go.mod template**
6. **Create main.go template for generated apps**
7. **Create tracks.yaml configuration template**
8. **Create basic directory structure templates**
9. **Add template validation logic**
10. **Write template rendering tests**
11. **Document template structure and variables**

## Dependencies

### Prerequisites

- Epic 1 (CLI Infrastructure) - need CLI to test template rendering
- Go 1.25+ with embed support

### Blocks

- Epic 3 (Project Generation) - can't generate without templates
- Epic 4 (Tooling) - tooling templates depend on this system

## Acceptance Criteria

- [ ] Template files successfully embedded in tracks binary
- [ ] Can render templates with variable substitution
- [ ] go.mod template generates valid Go module file
- [ ] main.go template creates runnable application entry point
- [ ] Template rendering handles missing variables gracefully
- [ ] Template tests cover all templates and edge cases
- [ ] Documentation explains template structure and variables
- [ ] Template validation catches invalid syntax before generation

## Technical Notes

### Embed Structure

```go
//go:embed templates/*
var templateFS embed.FS

type TemplateRenderer struct {
    fs embed.FS
}

func (r *TemplateRenderer) Render(name string, data map[string]interface{}) (string, error) {
    // Read template from embedded FS
    // Substitute variables
    // Return rendered content
}
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
│   ├── go.sum.tmpl
│   ├── tracks.yaml.tmpl
│   ├── README.md.tmpl
│   └── .gitignore.tmpl
└── embed.go
```

### Keep It Simple

Use Go's text/template for now. Don't overcomplicate with complex template logic. The goal is simple variable substitution.

## Testing Strategy

- Unit tests for template rendering with various variable combinations
- Tests for missing/invalid variables
- Tests that generated files are valid (go.mod parses, etc.)
- Integration test that embeds and renders all templates

## Next Epic

[Epic 3: Project Generation →](./3-project-generation.md)
