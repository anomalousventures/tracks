// Package template provides a template rendering engine for generating project files.
//
// The template system uses Go's embed.FS to bundle template files into the binary,
// eliminating external dependencies and ensuring templates are always available at runtime.
// Templates use Go's text/template syntax for variable substitution.
//
// # Architecture
//
// The template system consists of three main components:
//
//  1. Renderer interface - defines template rendering operations
//  2. TemplateData struct - provides variables to templates
//  3. Embedded templates - .tmpl files bundled via embed.FS
//
// Templates are embedded from internal/templates/project/ and rendered using
// the Renderer interface with TemplateData for variable substitution.
//
// # Basic Usage
//
// Create a renderer and render a template:
//
//	import "github.com/anomalousventures/tracks/internal/templates"
//
//	renderer := template.NewRenderer(templates.FS)
//	data := template.TemplateData{
//	    ModuleName:  "github.com/user/myapp",
//	    ProjectName: "myapp",
//	    GoVersion:   "1.25",
//	}
//
//	content, err := renderer.Render("go.mod.tmpl", data)
//	if err != nil {
//	    // handle error
//	}
//
// # Rendering to Files
//
// Render directly to a file with automatic directory creation:
//
//	outputPath := filepath.Join(projectDir, "go.mod")
//	err := renderer.RenderToFile("go.mod.tmpl", data, outputPath)
//
// # Template Variables
//
// All templates have access to TemplateData fields:
//
//	{{.ModuleName}}  - Go module path (e.g., github.com/user/myapp)
//	{{.ProjectName}} - Project name (e.g., myapp)
//	{{.DBDriver}}    - Database driver (go-libsql, sqlite3, postgres)
//	{{.GoVersion}}   - Go version (e.g., 1.25)
//	{{.Year}}        - Current year for copyright
//
// # Adding New Templates
//
// To add a new template:
//
//  1. Create a .tmpl file in internal/templates/project/
//  2. Use {{.VariableName}} for variable substitution
//  3. Add tests in templates_test.go
//  4. Template path structure is preserved in output (cmd/server/main.go.tmpl → cmd/server/main.go)
//
// # Error Handling
//
// The package provides two error types:
//
//	TemplateError     - wraps errors from file I/O and rendering
//	ValidationError   - reports template syntax errors
//
// Both implement error unwrapping for error chain inspection.
//
// # Cross-Platform Support
//
// The package uses filepath for OS-specific paths and path for embed.FS paths,
// ensuring correct behavior on Windows, macOS, and Linux.
package template
