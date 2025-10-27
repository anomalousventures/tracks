package template

import (
	"embed"
	"io/fs"
)

// Renderer handles template rendering with variable substitution.
// It provides methods to render templates from an embedded filesystem,
// write rendered templates to files, and validate template syntax.
type Renderer interface {
	// Render renders a template by name with the given data and returns the result as a string.
	// The template name should include the .tmpl extension (e.g., "go.mod.tmpl").
	// Returns an error if the template does not exist or cannot be parsed/executed.
	Render(name string, data TemplateData) (string, error)

	// RenderToFile renders a template and writes it to the specified output path.
	// The output path should NOT include the .tmpl extension (e.g., "project/go.mod").
	// Parent directories will be created automatically if they don't exist.
	// Returns an error if rendering fails or the file cannot be written.
	RenderToFile(templateName string, data TemplateData, outputPath string) error

	// Validate checks if a template exists and has valid syntax.
	// Returns nil if the template is valid, or an error describing the problem.
	Validate(name string) error
}

// TemplateRenderer implements Renderer using Go's embed.FS.
// It reads templates from an embedded filesystem and renders them using text/template.
type TemplateRenderer struct {
	fs embed.FS
}

// NewRenderer creates a new TemplateRenderer that reads templates from the given embed.FS.
// The embedded filesystem should contain template files with .tmpl extension.
func NewRenderer(fs embed.FS) Renderer {
	return &TemplateRenderer{fs: fs}
}

// Render renders a template by name with the given data and returns the result as a string.
// This is a stub implementation that will be completed in a later task.
func (r *TemplateRenderer) Render(name string, data TemplateData) (string, error) {
	return "", &TemplateError{Template: name, Err: fs.ErrNotExist}
}

// RenderToFile renders a template and writes it to the specified output path.
// This is a stub implementation that will be completed in a later task.
func (r *TemplateRenderer) RenderToFile(templateName string, data TemplateData, outputPath string) error {
	return &TemplateError{Template: templateName, Err: fs.ErrNotExist}
}

// Validate checks if a template exists and has valid syntax.
// This is a stub implementation that will be completed in a later task.
func (r *TemplateRenderer) Validate(name string) error {
	return &TemplateError{Template: name, Err: fs.ErrNotExist}
}
