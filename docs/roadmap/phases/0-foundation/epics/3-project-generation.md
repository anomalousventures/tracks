# Epic 3: Project Generation

[← Back to Phase 0](../0-foundation.md) | [← Epic 2](./2-template-engine.md) | [Epic 4 →](./4-generated-tooling.md)

## Overview

Implement the `tracks new` command to generate complete project structures. This is the core value proposition of Phase 0 - developers can create a new Go web application with a single command. This epic brings together the CLI infrastructure and template system to create functional projects.

## Goals

- Working `tracks new` command that generates complete projects
- Database driver selection during project creation
- Clean directory structure matching PRD specifications
- Valid Go module initialization
- Optional git repository initialization
- Robust error handling for edge cases

## Scope

### In Scope

- `tracks new <project>` command implementation
- Database driver flag (`--db go-libsql|sqlite3|postgres`)
- Module name flag (`--module github.com/user/app`)
- Git initialization flag (`--no-git`)
- Complete directory structure creation
- Configuration file generation (go.mod, tracks.yaml, .env.example)
- Validation (project name, module name, existing directories)
- User-friendly error messages and prompts

### Out of Scope

- Interactive prompts (TUI mode) - defer to later phase
- Custom template selection - use defaults only
- Database schema generation - that's Phase 2
- Full application code - just scaffold structure
- Migration files - Phase 2
- Test scaffolding - Phase 2 and beyond

## Task Breakdown

The following tasks will become GitHub issues:

1. **Implement `tracks new` command with Cobra**
2. **Add --db flag for database driver selection**
3. **Add --module flag for custom module names**
4. **Add --no-git flag to skip git initialization**
5. **Implement directory structure generation**
6. **Generate go.mod with correct module name**
7. **Generate tracks.yaml configuration file**
8. **Generate .env.example with sensible defaults**
9. **Implement optional git initialization**
10. **Add project name validation**
11. **Add existing directory detection and handling**
12. **Add module name validation**
13. **Implement post-generation summary and next steps**
14. **Write integration tests for project generation**
15. **Add cross-platform path handling**
16. **Document `tracks new` command and all flags**

## Dependencies

### Prerequisites

- Epic 1 (CLI Infrastructure) - need command framework
- Epic 2 (Template Engine) - need templates to render
- Go module system understanding

### Blocks

- Epic 4 (Tooling) - generates projects that tooling will enhance
- All future features that extend project generation

## Acceptance Criteria

- [ ] `tracks new myapp` creates valid project structure
- [ ] `tracks new myapp --db postgres` sets PostgreSQL driver
- [ ] `tracks new myapp --module github.com/me/app` uses custom module
- [ ] `tracks new myapp --no-git` skips git initialization
- [ ] Generated go.mod has correct module name and Go version
- [ ] Generated project directory matches PRD structure
- [ ] Generated tracks.yaml has sensible defaults
- [ ] Command fails gracefully if directory already exists
- [ ] Command validates project name (no spaces, special chars)
- [ ] Command validates module name (valid Go import path)
- [ ] Integration test generates project and runs `go mod download`
- [ ] Works on Linux, macOS, and Windows
- [ ] Post-generation shows helpful next steps

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

### Validation Rules

- **Project name:** lowercase, alphanumeric, hyphens/underscores allowed, no spaces
- **Module name:** valid Go import path format
- **Directory:** must not exist, or be empty

### Error Handling

Provide clear, actionable error messages:

- "Directory 'myapp' already exists. Use a different name or remove the directory."
- "Invalid project name 'My App'. Use lowercase letters, numbers, hyphens, and underscores."
- "Invalid module name 'not a valid module'. Must be a valid Go import path."

## Testing Strategy

- Unit tests for validation logic
- Integration tests that generate projects and verify structure
- Test all flag combinations
- Test error cases (existing directory, invalid names)
- Verify generated project can run `go mod download`

## Next Epic

[Epic 4: Generated Project Tooling →](./4-generated-tooling.md)
