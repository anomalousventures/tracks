package interfaces

// Validator validates project configuration values.
//
// This interface is owned by the CLI commands package, following ADR-002:
// interfaces are defined by consumers, not providers. The validation
// implementation lives in internal/generator/validator_impl.go.
//
// This pattern prevents import cycles and enables proper dependency inversion:
// CLI (high-level) defines interface, validation (low-level) implements it.
//
// Example usage:
//
//	logger := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
//	validator := generator.NewValidator(logger)
//	if err := validator.ValidateProjectName("my-app"); err != nil {
//	    return err
//	}
type Validator interface {
	// ValidateProjectName checks if project name is valid.
	// Rules: lowercase, alphanumeric, hyphens/underscores allowed, no spaces.
	ValidateProjectName(name string) error

	// ValidateModulePath checks if module path is a valid Go import path.
	// Must follow Go module naming conventions.
	ValidateModulePath(path string) error

	// ValidateDirectory checks if target directory is valid for project creation.
	// Rules: directory doesn't exist or is empty, parent directory exists and is writable.
	ValidateDirectory(path string) error

	// ValidateDatabaseDriver checks if the database driver is supported.
	// Valid drivers: go-libsql, sqlite3, postgres
	ValidateDatabaseDriver(driver string) error
}
