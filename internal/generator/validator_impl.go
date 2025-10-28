package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anomalousventures/tracks/internal/cli"
	"github.com/go-playground/validator/v10"
)

const (
	maxProjectNameLength = 100
	maxModulePathLength  = 300
)

var (
	projectNameRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)
	modulePathRegex  = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
)

// validatorImpl implements the Validator interface using go-playground/validator.
type validatorImpl struct {
	validate *validator.Validate
}

// NewValidator creates a new Validator with custom validation rules.
func NewValidator() Validator {
	v := validator.New()

	if err := v.RegisterValidation("project_name", func(fl validator.FieldLevel) bool {
		name := fl.Field().String()
		if len(name) == 0 || len(name) > maxProjectNameLength {
			return false
		}
		return projectNameRegex.MatchString(name)
	}); err != nil {
		panic(fmt.Sprintf("failed to register project_name validator: %v", err))
	}

	if err := v.RegisterValidation("module_path", func(fl validator.FieldLevel) bool {
		path := fl.Field().String()
		if path == "" || len(path) > maxModulePathLength {
			return false
		}

		if strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
			return false
		}

		if !strings.Contains(path, "/") {
			return false
		}

		return modulePathRegex.MatchString(path)
	}); err != nil {
		panic(fmt.Sprintf("failed to register module_path validator: %v", err))
	}

	return &validatorImpl{validate: v}
}

// ValidateProjectName validates that a project name follows the required format:
// lowercase alphanumeric with hyphens and underscores, max 100 characters.
func (v *validatorImpl) ValidateProjectName(name string) error {
	cfg := ProjectConfig{
		ProjectName:    name,
		ModulePath:     "placeholder",
		DatabaseDriver: "go-libsql",
		OutputPath:     "placeholder",
	}

	if err := v.validate.StructPartial(cfg, "ProjectName"); err != nil {
		if len(name) > maxProjectNameLength {
			return &ValidationError{
				Field:   "project_name",
				Value:   name,
				Message: fmt.Sprintf("must be %d characters or less", maxProjectNameLength),
				Err:     ErrInvalidProjectName,
			}
		}
		return &ValidationError{
			Field:   "project_name",
			Value:   name,
			Message: "must be lowercase alphanumeric with hyphens/underscores",
			Err:     ErrInvalidProjectName,
		}
	}

	return nil
}

// ValidateModulePath validates that a module path is a valid Go import path.
// Must contain domain and path, cannot start/end with slash, max 300 characters.
func (v *validatorImpl) ValidateModulePath(path string) error {
	cfg := ProjectConfig{
		ProjectName:    "placeholder",
		ModulePath:     path,
		DatabaseDriver: "go-libsql",
		OutputPath:     "placeholder",
	}

	if err := v.validate.StructPartial(cfg, "ModulePath"); err != nil {
		if path == "" {
			return &ValidationError{
				Field:   "module_path",
				Value:   path,
				Message: "cannot be empty",
				Err:     ErrInvalidModulePath,
			}
		}
		if len(path) > maxModulePathLength {
			return &ValidationError{
				Field:   "module_path",
				Value:   path,
				Message: fmt.Sprintf("must be %d characters or less", maxModulePathLength),
				Err:     ErrInvalidModulePath,
			}
		}
		if !strings.Contains(path, "/") {
			return &ValidationError{
				Field:   "module_path",
				Value:   path,
				Message: "must contain domain and path (e.g., github.com/user/project)",
				Err:     ErrInvalidModulePath,
			}
		}
		if strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
			return &ValidationError{
				Field:   "module_path",
				Value:   path,
				Message: "cannot start or end with slash",
				Err:     ErrInvalidModulePath,
			}
		}
		return &ValidationError{
			Field:   "module_path",
			Value:   path,
			Message: "must be valid Go import path",
			Err:     ErrInvalidModulePath,
		}
	}

	return nil
}

// ValidateDirectory checks if the target directory is valid for project creation.
// The directory must either not exist, or exist and be empty. The parent directory
// must exist and be writable.
func (v *validatorImpl) ValidateDirectory(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return &ValidationError{
				Field:   "output_path",
				Value:   path,
				Message: "path exists but is not a directory",
				Err:     ErrDirectoryExists,
			}
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("read directory: %w", err)
		}

		if len(entries) > 0 {
			return &ValidationError{
				Field:   "output_path",
				Value:   path,
				Message: "directory must be empty",
				Err:     ErrDirectoryExists,
			}
		}

		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("check directory: %w", err)
	}

	parent := filepath.Dir(path)
	if _, err := os.Stat(parent); os.IsNotExist(err) {
		return &ValidationError{
			Field:   "output_path",
			Value:   path,
			Message: "parent directory does not exist",
			Err:     ErrDirectoryNotWritable,
		}
	}

	testFile := filepath.Join(parent, ".tracks_write_test")
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		return &ValidationError{
			Field:   "output_path",
			Value:   path,
			Message: "parent directory is not writable",
			Err:     ErrDirectoryNotWritable,
		}
	}

	if err := os.Remove(testFile); err != nil {
		logger := cli.GetLogger(context.Background())
		logger.Warn().
			Err(err).
			Str("path", testFile).
			Msg("failed to cleanup validation test file")
	}

	return nil
}

// ValidateDatabaseDriver checks if the database driver is supported.
// Valid drivers: go-libsql, sqlite3, postgres (case-sensitive).
func (v *validatorImpl) ValidateDatabaseDriver(driver string) error {
	cfg := ProjectConfig{
		ProjectName:    "placeholder",
		ModulePath:     "placeholder",
		DatabaseDriver: driver,
		OutputPath:     "placeholder",
	}

	if err := v.validate.StructPartial(cfg, "DatabaseDriver"); err != nil {
		return &ValidationError{
			Field:   "database_driver",
			Value:   driver,
			Message: "must be one of: go-libsql, sqlite3, postgres",
			Err:     ErrInvalidDatabaseDriver,
		}
	}

	return nil
}
