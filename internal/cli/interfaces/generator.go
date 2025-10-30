package interfaces

import "context"

// ProjectGenerator generates new Tracks projects from templates.
//
// This interface is owned by the CLI commands package, following ADR-002:
// interfaces are defined by consumers, not providers. The generator
// implementation will live in internal/generator/.
//
// This pattern prevents import cycles and enables proper dependency inversion:
// CLI (high-level) defines interface, generator (low-level) implements it.
//
// Note: The config parameter uses 'any' to avoid import cycles. The actual
// implementation uses generator.ProjectConfig. This follows the pattern of
// accepting concrete types from the provider package without importing it.
//
// Example usage:
//
//	gen := generator.NewProjectGenerator()
//	cfg := generator.ProjectConfig{
//	    Name: "myapp",
//	    ModulePath: "github.com/user/myapp",
//	}
//	if err := gen.Generate(ctx, cfg); err != nil {
//	    return err
//	}
type ProjectGenerator interface {
	// Generate creates a new project with the given configuration.
	// It creates the directory structure, renders templates, and optionally
	// initializes a git repository. Returns an error if validation fails
	// or if any file system operations fail.
	//
	// The cfg parameter should be a generator.ProjectConfig value.
	Generate(ctx context.Context, cfg any) error

	// Validate checks if the configuration is valid without creating any files.
	// This allows callers to check for errors before attempting generation.
	//
	// The cfg parameter should be a generator.ProjectConfig value.
	Validate(cfg any) error
}
