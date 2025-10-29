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
// Data Parameter Design:
//
// The data parameter uses 'any' instead of a specific type (like TemplateData)
// to provide flexibility for future template rendering scenarios. While the
// current implementation exclusively uses template.TemplateData, the 'any' type
// allows the interface to support:
//   - Future data structure variations without interface changes
//   - Test scenarios with mock or simplified data structures
//   - Custom data types for specialized template rendering use cases
//
// Implementations are expected to document their supported data types.
// The standard implementation (template.NewRenderer) accepts template.TemplateData.
//
// Example usage:
//
//	renderer := template.NewRenderer(templateFS)
//	data := template.TemplateData{
//	    ModuleName:  "github.com/user/myapp",
//	    ProjectName: "myapp",
//	    GoVersion:   "1.25",
//	}
//	content, err := renderer.Render("go.mod.tmpl", data)
//	if err != nil {
//	    return fmt.Errorf("failed to render template: %w", err)
//	}
//
// See also: internal/generator/template package for the concrete implementation.
type TemplateRenderer interface {
	// Render renders a template by name with the given data and returns the result as a string.
	// The template name should include the .tmpl extension (e.g., "go.mod.tmpl").
	//
	// The data parameter accepts any type, but callers should use template.TemplateData
	// for standard project generation. See interface documentation for details on the
	// design rationale for using 'any'.
	//
	// Returns an error if the template does not exist or cannot be parsed/executed.
	Render(name string, data any) (string, error)

	// RenderToFile renders a template and writes it to the specified output path.
	// The output path should NOT include the .tmpl extension (e.g., "project/go.mod").
	// Parent directories will be created automatically if they don't exist.
	//
	// The data parameter accepts any type, but callers should use template.TemplateData
	// for standard project generation. See interface documentation for details on the
	// design rationale for using 'any'.
	//
	// Returns an error if rendering fails or the file cannot be written.
	RenderToFile(templateName string, data any, outputPath string) error

	// Validate checks if a template exists and has valid syntax.
	// Returns nil if the template is valid, or an error describing the problem.
	Validate(name string) error
}
