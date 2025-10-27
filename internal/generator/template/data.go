package template

// TemplateData contains all variables available to templates during rendering.
// This struct defines the complete schema of data that can be used in template files.
type TemplateData struct {
	// ModuleName is the full Go module path for the generated project.
	// Example: "github.com/user/myapp"
	ModuleName string

	// ProjectName is the short name of the project, typically the last segment of the module path.
	// Example: "myapp"
	ProjectName string

	// DBDriver specifies the database driver to use in the generated project.
	// Valid values: "go-libsql", "sqlite3", "postgres"
	DBDriver string

	// GoVersion specifies the Go version to use in the generated project's go.mod file.
	// Example: "1.25"
	GoVersion string

	// Year is the current year, used for copyright notices in generated files.
	// Example: 2025
	Year int
}
