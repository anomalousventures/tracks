package interfaces

import "context"

// Validator validates project configuration values.
//
// Interface defined by consumer per ADR-002 to avoid import cycles.
// Context parameter enables request-scoped logger access per ADR-003.
type Validator interface {
	ValidateProjectName(ctx context.Context, name string) error
	ValidateModulePath(ctx context.Context, path string) error
	ValidateDirectory(ctx context.Context, path string) error
	ValidateDatabaseDriver(ctx context.Context, driver string) error
	ValidateEnvPrefix(ctx context.Context, prefix string) error
}
