package template

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
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
// It reads the template from the embedded FS, parses it using text/template,
// executes it with the provided data, and returns the rendered string.
func (r *TemplateRenderer) Render(name string, data TemplateData) (string, error) {
	path := filepath.Join("project", name)
	content, err := fs.ReadFile(r.fs, path)
	if err != nil {
		return "", &TemplateError{Template: name, Err: err}
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return "", &TemplateError{Template: name, Err: err}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", &TemplateError{Template: name, Err: err}
	}

	return buf.String(), nil
}

// RenderToFile renders a template and writes it to the specified output path.
// It creates parent directories if they don't exist and writes the rendered content to the file.
// Uses cross-platform path handling with filepath package.
func (r *TemplateRenderer) RenderToFile(templateName string, data TemplateData, outputPath string) error {
	content, err := r.Render(templateName, data)
	if err != nil {
		return err
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &TemplateError{Template: templateName, Err: fmt.Errorf("failed to create directory: %w", err)}
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return &TemplateError{Template: templateName, Err: fmt.Errorf("failed to write file: %w", err)}
	}

	return nil
}

// Validate checks if a template exists and has valid syntax.
// This is a stub implementation that will be completed in a later task.
func (r *TemplateRenderer) Validate(name string) error {
	return &TemplateError{Template: name, Err: fs.ErrNotExist}
}
