# Template System

The template package provides a rendering engine for generating project files from embedded templates. It uses Go's `embed.FS` to bundle template files into the binary and `text/template` for variable substitution.

## Architecture

### Components

The template system consists of three main parts:

1. **Renderer Interface** - Defines operations for rendering templates
2. **TemplateData** - Schema for variables available to all templates
3. **Embedded Templates** - `.tmpl` files bundled via `embed.FS`

### Flow

```text
embed.FS (templates) → Renderer → text/template → Rendered Output
                          ↓
                     TemplateData (variables)
```

Templates are stored in `internal/templates/project/` and embedded at compile time. The Renderer reads templates from the embedded filesystem, executes them with TemplateData, and produces output files.

## Variable Schema

All templates have access to `TemplateData` fields:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `ModuleName` | `string` | Go module path | `github.com/user/myapp` |
| `ProjectName` | `string` | Project name (last segment of module path) | `myapp` |
| `DBDriver` | `string` | Database driver to use | `go-libsql`, `sqlite3`, `postgres` |
| `GoVersion` | `string` | Go version for go.mod | `1.25` |
| `Year` | `int` | Current year for copyright notices | `2025` |

### Example Template Usage

In a template file (`go.mod.tmpl`):

```go
module {{.ModuleName}}

go {{.GoVersion}}
```

With TemplateData:

```go
data := TemplateData{
    ModuleName: "github.com/user/myapp",
    GoVersion:  "1.25",
}
```

Renders to:

```text
module github.com/user/myapp

go 1.25
```

## Template File Naming

Templates use the `.tmpl` extension and preserve directory structure:

| Template Path | Output Path |
|---------------|-------------|
| `go.mod.tmpl` | `go.mod` |
| `.gitignore.tmpl` | `.gitignore` |
| `cmd/server/main.go.tmpl` | `cmd/server/main.go` |
| `tracks.yaml.tmpl` | `tracks.yaml` |

The `.tmpl` extension is stripped during rendering, and nested paths are preserved in the output.

## Using the Renderer

### Creating a Renderer

```go
import (
    "github.com/anomalousventures/tracks/internal/generator/template"
    "github.com/anomalousventures/tracks/internal/templates"
)

renderer := template.NewRenderer(templates.FS)
```

### Rendering to String

```go
data := template.TemplateData{
    ModuleName:  "github.com/user/myapp",
    ProjectName: "myapp",
    GoVersion:   "1.25",
}

content, err := renderer.Render("go.mod.tmpl", data)
if err != nil {
    // handle error
}
```

### Rendering to File

The `RenderToFile` method automatically creates parent directories:

```go
outputPath := filepath.Join(projectDir, "go.mod")
err := renderer.RenderToFile("go.mod.tmpl", data, outputPath)
if err != nil {
    // handle error
}
```

For nested paths:

```go
mainPath := filepath.Join(projectDir, "cmd", "server", "main.go")
err := renderer.RenderToFile("cmd/server/main.go.tmpl", data, mainPath)
// Creates cmd/ and cmd/server/ directories automatically
```

### Validating Templates

Check if a template exists and has valid syntax:

```go
err := renderer.Validate("go.mod.tmpl")
if err != nil {
    // Template is missing or has syntax errors
}
```

## Adding New Templates

To add a new template to the system:

### 1. Create Template File

Create a `.tmpl` file in `internal/templates/project/`:

```bash
# Root-level template
internal/templates/project/myfile.txt.tmpl

# Nested template
internal/templates/project/config/settings.yaml.tmpl
```

### 2. Use Template Variables

Use `{{.FieldName}}` for variable substitution:

```yaml
# settings.yaml.tmpl
app_name: {{.ProjectName}}
version: {{.GoVersion}}
database:
  driver: {{.DBDriver}}
```

### 3. Write Tests

Add tests in `templates_test.go`:

```go
func TestMyFileTemplate(t *testing.T) {
    renderer := NewRenderer(templates.FS)

    data := TemplateData{
        ProjectName: "testapp",
        GoVersion:   "1.25",
        DBDriver:    "sqlite3",
    }

    result, err := renderer.Render("myfile.txt.tmpl", data)
    require.NoError(t, err)
    assert.Contains(t, result, "testapp")
}
```

### 4. Update Production Templates List

If the template is part of standard project generation, add it to `productionTemplates` in `integration_test.go`.

## Cross-Platform Path Handling

The package uses two path packages for correct cross-platform behavior:

- **`path`** - For embed.FS paths (always forward slashes)
- **`filepath`** - For OS-specific file paths (uses OS separators)

### Embed Paths (internal use)

```go
embedPath := path.Join("project", templateName)  // Always uses /
content, err := fs.ReadFile(r.fs, embedPath)
```

### File System Paths (user-facing)

```go
outputPath := filepath.Join(projectDir, "cmd", "server", "main.go")
// Windows: project\cmd\server\main.go
// Unix:    project/cmd/server/main.go
```

This ensures templates work correctly on Windows, macOS, and Linux.

## Error Handling

The package provides two error types:

### TemplateError

Wraps errors from file I/O and rendering operations:

```go
_, err := renderer.Render("nonexistent.tmpl", data)
if terr, ok := err.(*template.TemplateError); ok {
    fmt.Println("Template:", terr.Template)  // "nonexistent.tmpl"
    fmt.Println("Cause:", terr.Err)          // underlying error
}
```

### ValidationError

Reports template syntax errors:

```go
err := renderer.Validate("bad-syntax.tmpl")
if verr, ok := err.(*template.ValidationError); ok {
    fmt.Println("Template:", verr.Template)  // "bad-syntax.tmpl"
    fmt.Println("Message:", verr.Message)    // syntax error details
}
```

Both error types support error unwrapping:

```go
errors.Unwrap(err)  // Get underlying error
errors.Is(err, fs.ErrNotExist)  // Check for specific errors
```

## Testing Templates

### Unit Tests

Test individual templates with various data combinations:

```go
func TestTemplateWithEmptyData(t *testing.T) {
    renderer := NewRenderer(templates.FS)
    result, err := renderer.Render("template.tmpl", TemplateData{})
    require.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

### Integration Tests

Test that all templates work together:

```go
func TestAllTemplatesRender(t *testing.T) {
    tmpDir := t.TempDir()
    renderer := NewRenderer(templates.FS)

    data := TemplateData{
        ModuleName:  "github.com/test/app",
        ProjectName: "app",
        GoVersion:   "1.25",
    }

    for _, tmpl := range productionTemplates {
        outputName := strings.TrimSuffix(tmpl, ".tmpl")
        outputPath := filepath.Join(tmpDir, outputName)

        err := renderer.RenderToFile(tmpl, data, outputPath)
        require.NoError(t, err)

        _, err = os.Stat(outputPath)
        require.NoError(t, err, "file should exist")
    }
}
```

### Cross-Platform Tests

Verify templates work on all operating systems:

```go
func TestCrossPlatform(t *testing.T) {
    tmpDir := t.TempDir()
    renderer := NewRenderer(templates.FS)

    // Test nested directory creation
    nestedPath := filepath.Join(tmpDir, "cmd", "server", "main.go")
    err := renderer.RenderToFile("cmd/server/main.go.tmpl", data, nestedPath)
    require.NoError(t, err)

    _, err = os.Stat(nestedPath)
    require.NoError(t, err)
}
```

## Best Practices

### Template Design

1. **Keep templates simple** - Use variable substitution, avoid complex logic
2. **Make templates self-contained** - Each template should be independently renderable
3. **Use descriptive variable names** - `{{.ProjectName}}` not `{{.Name}}`
4. **Preserve formatting** - Templates should produce properly formatted output

### Variable Naming

1. **Use PascalCase** - `ModuleName`, `ProjectName`
2. **Be specific** - `DBDriver` not `Driver`
3. **Document in TemplateData** - Add godoc comments for new fields

### Testing

1. **Test with empty data** - Ensure templates don't crash with zero values
2. **Test with different drivers** - Verify `DBDriver` variations work
3. **Test cross-platform** - Use `filepath.Join` in tests
4. **Test edge cases** - Special characters, long names, etc.

### File Organization

1. **Mirror output structure** - Template path = output path + `.tmpl`
2. **Group related templates** - Use subdirectories for logical grouping
3. **Use clear names** - `main.go.tmpl` not `m.tmpl`

## Extending the System

### Adding New Variables

To add a new variable to TemplateData:

1. **Add field to struct** in `data.go` (see example below)
2. **Update tests** to include the new field
3. **Update documentation** (this README and godoc)
4. **Use in templates**: `{{.Description}}`

Example structure in `data.go`:

```go
type TemplateData struct {
    // ... existing fields ...

    // Description is the project description for README
    Description string
}
```

### Custom Renderer Implementations

The `Renderer` interface allows custom implementations:

```go
type Renderer interface {
    Render(name string, data TemplateData) (string, error)
    RenderToFile(templateName string, data TemplateData, outputPath string) error
    Validate(name string) error
}
```

Example custom renderer:

```go
type CachingRenderer struct {
    base  Renderer
    cache map[string]string
}

func (r *CachingRenderer) Render(name string, data TemplateData) (string, error) {
    if cached, ok := r.cache[name]; ok {
        return cached, nil
    }

    result, err := r.base.Render(name, data)
    if err == nil {
        r.cache[name] = result
    }
    return result, err
}
```

## Godoc

View full API documentation:

```bash
go doc github.com/anomalousventures/tracks/internal/generator/template
go doc github.com/anomalousventures/tracks/internal/generator/template.Renderer
go doc github.com/anomalousventures/tracks/internal/generator/template.TemplateData
```

Or visit: https://pkg.go.dev/github.com/anomalousventures/tracks/internal/generator/template
