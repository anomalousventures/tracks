package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	// Postgres driver imported for direct CLI connections.
	// SQLite-based drivers (sqlite3, go-libsql) have symbol conflicts when linked together,
	// so SQLite projects use generated make commands instead (avoids CGO conflicts).
	_ "github.com/lib/pq"
)

// Manager implements the DatabaseManager interface for CLI database operations.
type Manager struct {
	driver      string
	databaseURL string
	db          *sql.DB
	envLoaded   bool
}

func NewManager(driver string) *Manager {
	return &Manager{
		driver: driver,
	}
}

func (m *Manager) LoadEnv(ctx context.Context, projectDir string) error {
	logger := zerolog.Ctx(ctx)

	envPath := filepath.Join(projectDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			logger.Debug().Err(err).Str("path", envPath).Msg("failed to load .env file")
			return fmt.Errorf("failed to load .env file: %w", err)
		}
		logger.Debug().Str("path", envPath).Msg("loaded .env file")
	} else {
		logger.Debug().Str("path", envPath).Msg(".env file not found, using environment variables only")
	}

	m.databaseURL = os.Getenv("DATABASE_URL")
	m.envLoaded = true

	return nil
}

func (m *Manager) GetDatabaseURL() string {
	return m.databaseURL
}

func (m *Manager) GetDriver() string {
	return m.driver
}

func (m *Manager) Connect(ctx context.Context) (*sql.DB, error) {
	if !m.envLoaded {
		return nil, ErrEnvNotLoaded
	}

	if m.db != nil {
		return nil, ErrAlreadyConnected
	}

	if m.databaseURL == "" {
		return nil, ErrDatabaseURLNotSet
	}

	logger := zerolog.Ctx(ctx)

	driverName, err := m.sqlDriverName()
	if err != nil {
		return nil, err
	}

	logger.Debug().
		Str("driver", driverName).
		Msg("connecting to database")

	db, err := sql.Open(driverName, m.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := ctx.Err(); err != nil {
		db.Close()
		return nil, fmt.Errorf("context cancelled before ping: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	m.db = db
	logger.Debug().Msg("database connection established")

	return db, nil
}

func (m *Manager) Close() error {
	if m.db == nil {
		return ErrNotConnected
	}

	err := m.db.Close()
	m.db = nil
	return err
}

func (m *Manager) IsConnected() bool {
	return m.db != nil
}

func (m *Manager) sqlDriverName() (string, error) {
	switch m.driver {
	case "postgres":
		return "postgres", nil
	case "sqlite3", "go-libsql":
		return "", fmt.Errorf("%w: %s (use project's make commands for SQLite databases)", ErrUnsupportedDriver, m.driver)
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedDriver, m.driver)
	}
}
