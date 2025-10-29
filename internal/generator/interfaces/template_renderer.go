package interfaces

// TemplateRenderer handles template rendering with variable substitution.
//
// This interface follows ADR-002: Interface Placement in Consumer Packages.
// It is defined by the consumer (generator package) rather than the provider
// (template package), allowing the generator to depend on abstractions rather
// than concrete implementations.
//
// The interface provides methods to render templates from an embedded filesystem,
// write rendered templates to files, and validate template syntax.
//
// Example usage:
//
//	renderer := template.NewRenderer(templateFS)
//	content, err := renderer.Render("go.mod.tmpl", templateData)
//	if err != nil {
//	    return fmt.Errorf("failed to render template: %w", err)
//	}
//
// See also: internal/generator/template package for the concrete implementation.
type TemplateRenderer interface {
	// Render renders a template by name with the given data and returns the result as a string.
	// The template name should include the .tmpl extension (e.g., "go.mod.tmpl").
	// Returns an error if the template does not exist or cannot be parsed/executed.
	Render(name string, data any) (string, error)

	// RenderToFile renders a template and writes it to the specified output path.
	// The output path should NOT include the .tmpl extension (e.g., "project/go.mod").
	// Parent directories will be created automatically if they don't exist.
	// Returns an error if rendering fails or the file cannot be written.
	RenderToFile(templateName string, data any, outputPath string) error

	// Validate checks if a template exists and has valid syntax.
	// Returns nil if the template is valid, or an error describing the problem.
	Validate(name string) error
}
