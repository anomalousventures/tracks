package template

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/anomalousventures/tracks/internal/generator/interfaces"
)

// templateRenderer implements interfaces.TemplateRenderer using Go's embed.FS.
// It reads templates from an embedded filesystem and renders them using text/template.
//
// This implementation follows ADR-002 by implementing an interface defined by the consumer
// (generator package) rather than defining its own interface. This allows the generator
// to depend on abstractions rather than concrete implementations.
type templateRenderer struct {
	fs embed.FS
}

// NewRenderer creates a new template renderer that implements interfaces.TemplateRenderer.
// The provided embed.FS should contain template files in a "project" subdirectory.
func NewRenderer(fs embed.FS) interfaces.TemplateRenderer {
	return &templateRenderer{fs: fs}
}

func (r *templateRenderer) Render(name string, data any) (string, error) {
	embedPath := path.Join("project", name)
	content, err := fs.ReadFile(r.fs, embedPath)
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

func (r *templateRenderer) RenderToFile(templateName string, data any, outputPath string) error {
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

func (r *templateRenderer) Validate(name string) error {
	embedPath := path.Join("project", name)
	content, err := fs.ReadFile(r.fs, embedPath)
	if err != nil {
		return &TemplateError{Template: name, Err: err}
	}

	_, err = template.New(name).Parse(string(content))
	if err != nil {
		return &ValidationError{Template: name, Message: err.Error()}
	}

	return nil
}
