package interfaces

import "context"

// Validator validates project configuration values.
//
// This interface is owned by the CLI commands package, following ADR-002:
// interfaces are defined by consumers, not providers. The validation
// implementation lives in internal/validation/validator.go.
//
// Following ADR-003 (Context Propagation), all methods accept context.Context
// as the first parameter to enable logger access via cli.GetLogger(ctx).
//
// This pattern prevents import cycles and enables proper dependency inversion:
// CLI (high-level) defines interface, validation (low-level) implements it.
//
// Example usage:
//
//	ctx := cmd.Context()
//	validator := validation.NewValidator()
//	if err := validator.ValidateProjectName(ctx, "my-app"); err != nil {
//	    return err
//	}
type Validator interface {
	ValidateProjectName(ctx context.Context, name string) error
	ValidateModulePath(ctx context.Context, path string) error
	ValidateDirectory(ctx context.Context, path string) error
	ValidateDatabaseDriver(ctx context.Context, driver string) error
}
