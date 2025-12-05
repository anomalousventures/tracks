package interfaces

import (
	"context"
	"database/sql"
)

// DatabaseManager handles database connection and environment loading for CLI commands.
//
// Interface defined by consumer per ADR-002 to avoid import cycles.
// Context parameter enables request-scoped logger access per ADR-003.
type DatabaseManager interface {
	LoadEnv(ctx context.Context, projectDir string) error
	GetDatabaseURL() string
	GetDriver() string

	// Must call LoadEnv first.
	Connect(ctx context.Context) (*sql.DB, error)

	Close() error
	IsConnected() bool
}
