package generator

// Validator validates project configuration.
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
