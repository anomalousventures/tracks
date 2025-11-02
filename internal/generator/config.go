package generator

// ProjectConfig holds all configuration for generating a new project.
type ProjectConfig struct {
	ProjectName    string `json:"project_name" validate:"required,project_name"`
	ModulePath     string `json:"module_path" validate:"required,module_path"`
	DatabaseDriver string `json:"database_driver" validate:"required,oneof=go-libsql sqlite3 postgres"`
	InitGit        bool   `json:"init_git"`
	OutputPath     string `json:"output_path" validate:"required"`
}
