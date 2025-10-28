package generator

// ProjectConfig holds all configuration for generating a new project.
type ProjectConfig struct {
	// ProjectName is the name of the project.
	// Must be lowercase, alphanumeric with hyphens/underscores allowed, no spaces.
	ProjectName string `json:"project_name"`

	// ModulePath is the Go module import path (e.g., github.com/user/project).
	// Must be a valid Go import path.
	ModulePath string `json:"module_path"`

	// DatabaseDriver is the database driver to use.
	// Valid values: go-libsql, sqlite3, postgres
	DatabaseDriver string `json:"database_driver"`

	// InitGit indicates whether to initialize a git repository.
	InitGit bool `json:"init_git"`

	// OutputPath is the directory where the project will be created.
	// The project directory will be created as a subdirectory of this path.
	OutputPath string `json:"output_path"`
}
