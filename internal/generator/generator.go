package generator

import "context"

// ProjectGenerator creates new Tracks projects from templates.
type ProjectGenerator interface {
	// Generate creates a new project with the given configuration.
	// It creates the directory structure, renders templates, and optionally
	// initializes a git repository. Returns an error if validation fails
	// or if any file system operations fail.
	Generate(ctx context.Context, cfg ProjectConfig) error

	// Validate checks if the configuration is valid without creating any files.
	// This allows callers to check for errors before attempting generation.
	Validate(cfg ProjectConfig) error
}
