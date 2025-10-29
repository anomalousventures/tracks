package interfaces

// BuildInfo provides version metadata for the CLI.
type BuildInfo interface {
	GetVersion() string
	GetCommit() string
	GetDate() string
}
