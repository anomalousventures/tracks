package interfaces

import "context"

// ProjectDetector detects and reads Tracks project configuration.
//
// Interface defined by consumer per ADR-002 to avoid import cycles.
// Context parameter enables request-scoped logger access per ADR-003.
type ProjectDetector interface {
	// Detect finds a Tracks project starting from the given directory.
	// Searches upward through parent directories for .tracks.yaml.
	// Returns project info and the directory containing .tracks.yaml.
	// Returns an error if no .tracks.yaml is found.
	Detect(ctx context.Context, startDir string) (*TracksProject, string, error)

	// HasTemplUIConfig checks if .templui.json exists in the project directory.
	HasTemplUIConfig(ctx context.Context, projectDir string) bool
}

// TracksProject contains metadata from .tracks.yaml.
type TracksProject struct {
	Name       string
	ModulePath string
	DBDriver   string
}
