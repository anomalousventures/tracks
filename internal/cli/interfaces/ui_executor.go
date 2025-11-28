package interfaces

import "context"

// UIExecutor wraps templUI CLI commands for UI component management.
//
// Interface defined by consumer per ADR-002 to avoid import cycles.
// Context parameter enables request-scoped logger access per ADR-003.
type UIExecutor interface {
	// Version returns the templUI CLI version string.
	Version(ctx context.Context, projectDir string) (string, error)

	// Add installs one or more components into the project.
	// If force is true, overwrites existing components.
	Add(ctx context.Context, projectDir string, components []string, force bool) error

	// List returns available components from the templUI registry.
	List(ctx context.Context, projectDir string) ([]UIComponent, error)

	// Upgrade updates installed components to latest versions.
	Upgrade(ctx context.Context, projectDir string) error

	// IsAvailable checks if templui tool is installed and accessible.
	IsAvailable(ctx context.Context, projectDir string) bool
}

// UIComponent represents a templUI component.
type UIComponent struct {
	Name      string
	Category  string
	Installed bool
}
