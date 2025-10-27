package template

import (
	"errors"
	"strings"
	"testing"
)

// TestTemplateErrorImplementsError verifies TemplateError implements the error interface
func TestTemplateErrorImplementsError(t *testing.T) {
	var _ error = (*TemplateError)(nil)
}

func TestTemplateErrorMessage(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		err          error
		wantContains []string
	}{
		{
			name:         "basic error",
			templateName: "test.tmpl",
			err:          errors.New("file not found"),
			wantContains: []string{"template test.tmpl", "file not found"},
		},
		{
			name:         "empty template name",
			templateName: "",
			err:          errors.New("parse error"),
			wantContains: []string{"template ", "parse error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &TemplateError{
				Template: tt.templateName,
				Err:      tt.err,
			}

			msg := err.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(msg, want) {
					t.Errorf("TemplateError.Error() = %q, want to contain %q", msg, want)
				}
			}
		})
	}
}

func TestTemplateErrorUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	templateErr := &TemplateError{
		Template: "test.tmpl",
		Err:      originalErr,
	}

	unwrapped := templateErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	if !errors.Is(templateErr, originalErr) {
		t.Error("errors.Is() should return true for wrapped error")
	}
}

// TestValidationErrorImplementsError verifies ValidationError implements the error interface
func TestValidationErrorImplementsError(t *testing.T) {
	var _ error = (*ValidationError)(nil)
}

func TestValidationErrorMessage(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		field        string
		message      string
		wantContains []string
	}{
		{
			name:         "error with field",
			templateName: "test.tmpl",
			field:        "ModuleName",
			message:      "cannot be empty",
			wantContains: []string{"template test.tmpl", "field ModuleName", "cannot be empty"},
		},
		{
			name:         "error without field",
			templateName: "test.tmpl",
			field:        "",
			message:      "invalid syntax",
			wantContains: []string{"template test.tmpl", "invalid syntax"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ValidationError{
				Template: tt.templateName,
				Field:    tt.field,
				Message:  tt.message,
			}

			msg := err.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(msg, want) {
					t.Errorf("ValidationError.Error() = %q, want to contain %q", msg, want)
				}
			}
		})
	}
}

func TestValidationErrorWithoutField(t *testing.T) {
	err := &ValidationError{
		Template: "test.tmpl",
		Field:    "",
		Message:  "template validation failed",
	}

	msg := err.Error()
	if strings.Contains(msg, "field :") || strings.Contains(msg, "field  :") {
		t.Errorf("ValidationError.Error() should not include 'field' when Field is empty, got: %q", msg)
	}

	wantContains := []string{"template test.tmpl", "template validation failed"}
	for _, want := range wantContains {
		if !strings.Contains(msg, want) {
			t.Errorf("ValidationError.Error() = %q, want to contain %q", msg, want)
		}
	}
}
