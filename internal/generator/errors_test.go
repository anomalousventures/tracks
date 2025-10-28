package generator

import (
	"errors"
	"strings"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *ValidationError
		wantMsg string
	}{
		{
			name: "with value",
			err: &ValidationError{
				Field:   "project_name",
				Value:   "Invalid Name",
				Message: "must be lowercase",
			},
			wantMsg: "validation failed for project_name 'Invalid Name': must be lowercase",
		},
		{
			name: "without value",
			err: &ValidationError{
				Field:   "module_path",
				Message: "required field is empty",
			},
			wantMsg: "validation failed for module_path: required field is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	valErr := &ValidationError{
		Field:   "project_name",
		Value:   "test",
		Message: "test message",
		Err:     baseErr,
	}

	if !errors.Is(valErr, baseErr) {
		t.Error("expected ValidationError to wrap base error")
	}

	unwrapped := errors.Unwrap(valErr)
	if unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestValidationError_NoWrappedError(t *testing.T) {
	valErr := &ValidationError{
		Field:   "test",
		Message: "test",
		Err:     nil,
	}

	unwrapped := errors.Unwrap(valErr)
	if unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"invalid project name", ErrInvalidProjectName, "invalid project name"},
		{"invalid module path", ErrInvalidModulePath, "invalid module path"},
		{"directory exists", ErrDirectoryExists, "directory already exists"},
		{"directory not writable", ErrDirectoryNotWritable, "directory not writable"},
		{"invalid database driver", ErrInvalidDatabaseDriver, "invalid database driver"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.msg)
			}
		})
	}
}

func TestSentinelErrors_CanBeWrapped(t *testing.T) {
	wrapped := errors.Join(ErrInvalidProjectName, errors.New("contains spaces"))

	if !errors.Is(wrapped, ErrInvalidProjectName) {
		t.Error("expected wrapped error to match ErrInvalidProjectName")
	}

	msg := wrapped.Error()
	if !strings.Contains(msg, "invalid project name") {
		t.Errorf("wrapped error message %q should contain sentinel error message", msg)
	}
}
