package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/pressly/goose/v3"
)

type MigrationRunner struct {
	db       *sql.DB
	provider *goose.Provider
}

type MigrationStatus struct {
	Version   int64
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

type MigrationResult struct {
	Direction string // "up" or "down"
	Applied   []MigrationStatus
}

func NewMigrationRunner(db *sql.DB, driver string, migrationsDir string) (*MigrationRunner, error) {
	dialect, err := gooseDialect(driver)
	if err != nil {
		return nil, err
	}

	// Verify migrations directory exists
	info, err := os.Stat(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("migrations directory not found: %s", migrationsDir)
		}
		return nil, fmt.Errorf("failed to access migrations directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("migrations path is not a directory: %s", migrationsDir)
	}

	// Create filesystem from the migrations directory
	fsys := os.DirFS(migrationsDir)

	provider, err := goose.NewProvider(
		dialect,
		db,
		fsys,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration provider: %w", err)
	}

	return &MigrationRunner{
		db:       db,
		provider: provider,
	}, nil
}

func (r *MigrationRunner) Up(ctx context.Context, steps int) (*MigrationResult, error) {
	var results []*goose.MigrationResult
	var err error

	if steps <= 0 {
		results, err = r.provider.Up(ctx)
	} else {
		for appliedCount := 0; appliedCount < steps; appliedCount++ {
			result, upErr := r.provider.UpByOne(ctx)
			if upErr != nil {
				// ErrNoNextVersion means no more migrations to apply
				if upErr == goose.ErrNoNextVersion {
					break
				}
				err = upErr
				break
			}
			if result != nil {
				results = append(results, result)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &MigrationResult{
		Direction: "up",
		Applied:   gooseResultsToStatus(results),
	}, nil
}

func (r *MigrationRunner) Down(ctx context.Context, steps int) (*MigrationResult, error) {
	if steps <= 0 {
		steps = 1
	}

	var results []*goose.MigrationResult
	for i := 0; i < steps; i++ {
		result, err := r.provider.Down(ctx)
		if err != nil {
			if i == 0 {
				return nil, fmt.Errorf("rollback failed: %w", err)
			}
			// Return partial results: user should see which migrations succeeded
			// before the failure, enabling manual recovery decisions
			break
		}
		if result == nil {
			// No more migrations to roll back
			break
		}
		results = append(results, result)
	}

	return &MigrationResult{
		Direction: "down",
		Applied:   gooseResultsToStatus(results),
	}, nil
}

func (r *MigrationRunner) Status(ctx context.Context) ([]MigrationStatus, error) {
	sources := r.provider.ListSources()
	statuses, err := r.provider.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	result := make([]MigrationStatus, 0, len(sources))
	for _, status := range statuses {
		ms := MigrationStatus{
			Version: status.Source.Version,
			Name:    filepath.Base(status.Source.Path),
			Applied: status.State == goose.StateApplied,
		}
		if !status.AppliedAt.IsZero() {
			appliedAt := status.AppliedAt
			ms.AppliedAt = &appliedAt
		}
		result = append(result, ms)
	}

	return result, nil
}

func GetMigrationsDir(projectDir, driver string) string {
	return filepath.Join(projectDir, "internal", "db", "migrations", driver)
}

func gooseDialect(driver string) (goose.Dialect, error) {
	switch driver {
	case "postgres":
		return goose.DialectPostgres, nil
	case "sqlite3", "go-libsql":
		return goose.DialectSQLite3, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedDriver, driver)
	}
}

func gooseResultsToStatus(results []*goose.MigrationResult) []MigrationStatus {
	statuses := make([]MigrationStatus, 0, len(results))
	for _, r := range results {
		if r == nil || r.Source == nil {
			continue
		}
		statuses = append(statuses, MigrationStatus{
			Version: r.Source.Version,
			Name:    filepath.Base(r.Source.Path),
			Applied: true,
		})
	}
	return statuses
}

// Ensure MigrationRunner uses the fs.FS interface correctly
var _ fs.FS = os.DirFS(".")
