package template

import "fmt"

// TemplateError represents an error that occurred during template rendering.
// It wraps the underlying error with the template name for better context.
type TemplateError struct {
	// Template is the name of the template that caused the error
	Template string

	// Err is the underlying error
	Err error
}

// Error implements the error interface for TemplateError.
// It returns a formatted error message that includes the template name and underlying error.
func (e *TemplateError) Error() string {
	return fmt.Sprintf("template %s: %v", e.Template, e.Err)
}

// Unwrap returns the underlying error for error chain unwrapping.
func (e *TemplateError) Unwrap() error {
	return e.Err
}

// ValidationError represents an error that occurred during template validation.
// It includes the template name, the problematic field, and a descriptive message.
type ValidationError struct {
	// Template is the name of the template being validated
	Template string

	// Field is the name of the field that failed validation (empty if not field-specific)
	Field string

	// Message describes the validation failure
	Message string
}

// Error implements the error interface for ValidationError.
// It returns a formatted error message with template name, field, and message.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("template %s: field %s: %s", e.Template, e.Field, e.Message)
	}
	return fmt.Sprintf("template %s: %s", e.Template, e.Message)
}
