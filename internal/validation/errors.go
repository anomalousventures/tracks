package validation

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidProjectName is returned when project name is invalid.
	ErrInvalidProjectName = errors.New("invalid project name")

	// ErrInvalidModulePath is returned when module path is invalid.
	ErrInvalidModulePath = errors.New("invalid module path")

	// ErrDirectoryExists is returned when target directory already exists.
	ErrDirectoryExists = errors.New("directory already exists")

	// ErrDirectoryNotWritable is returned when directory is not writable.
	ErrDirectoryNotWritable = errors.New("directory not writable")

	// ErrInvalidDatabaseDriver is returned when database driver is not supported.
	ErrInvalidDatabaseDriver = errors.New("invalid database driver")
)

// ValidationError wraps validation failures with context.
type ValidationError struct {
	Field   string
	Value   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation failed for %s '%s': %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}
