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
	Generate(ctx context.Context, cfg any) error
	Validate(cfg any) error
}
