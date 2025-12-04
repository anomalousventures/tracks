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
	// LoadEnv loads environment variables from the project's .env file.
	// Environment variables already set take precedence over .env values.
	LoadEnv(ctx context.Context, projectDir string) error

	// GetDatabaseURL returns the DATABASE_URL from environment.
	// Returns empty string if not set.
	GetDatabaseURL() string

	// GetDriver returns the database driver name (sqlite3, postgres, go-libsql).
	GetDriver() string

	// Connect opens a database connection using the loaded configuration.
	// Must call LoadEnv before Connect.
	Connect(ctx context.Context) (*sql.DB, error)

	// Close closes the database connection if open.
	Close() error

	IsConnected() bool
}
