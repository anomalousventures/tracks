package commands

import (
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/internal/database"
	"github.com/spf13/cobra"
)

// DatabaseManagerFactory creates a DatabaseManager for a given driver.
type DatabaseManagerFactory func(driver string) interfaces.DatabaseManager

// DefaultDatabaseManagerFactory returns the production database manager factory.
func DefaultDatabaseManagerFactory() DatabaseManagerFactory {
	return func(driver string) interfaces.DatabaseManager {
		return database.NewManager(driver)
	}
}

type DBMigrateCommand struct {
	detector      interfaces.ProjectDetector
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
	newDBManager  DatabaseManagerFactory
}

func NewDBMigrateCommand(
	detector interfaces.ProjectDetector,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *DBMigrateCommand {
	return &DBMigrateCommand{
		detector:      detector,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
		newDBManager:  DefaultDatabaseManagerFactory(),
	}
}

// NewDBMigrateCommandWithFactory creates a DBMigrateCommand with a custom factory for testing.
func NewDBMigrateCommandWithFactory(
	detector interfaces.ProjectDetector,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
	newDBManager DatabaseManagerFactory,
) *DBMigrateCommand {
	return &DBMigrateCommand{
		detector:      detector,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
		newDBManager:  newDBManager,
	}
}

func (c *DBMigrateCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run pending database migrations",
		Long: `Run pending database migrations for your Tracks project.

Applies all pending migrations in order, or a specific number with --steps.
Use --dry-run to preview which migrations would be applied.

Note: This command only supports Postgres projects directly.
For SQLite/go-libsql projects, use: make migrate-up`,
		Example: `  # Run all pending migrations
  tracks db migrate

  # Run only 2 migrations
  tracks db migrate --steps 2

  # Preview migrations without running them
  tracks db migrate --dry-run`,
		RunE: c.runE,
	}

	cmd.Flags().IntP("steps", "n", 0, "Number of migrations to apply (0 = all pending)")
	cmd.Flags().Bool("dry-run", false, "Show migrations without applying them")

	return cmd
}

func (c *DBMigrateCommand) runE(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	steps, _ := cmd.Flags().GetInt("steps")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Detect project
	project, projectDir, err := c.detector.Detect(ctx, ".")
	if err != nil {
		return fmt.Errorf("not in a Tracks project directory (missing .tracks.yaml): %w", err)
	}

	// Check for supported driver
	if project.DBDriver != "postgres" {
		return fmt.Errorf("tracks db migrate only supports Postgres projects (found: %s). For SQLite/go-libsql projects, use: make migrate-up", project.DBDriver)
	}

	// Create database manager
	dbManager := c.newDBManager(project.DBDriver)
	if err := dbManager.LoadEnv(ctx, projectDir); err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	// Check for DATABASE_URL
	if dbManager.GetDatabaseURL() == "" {
		return fmt.Errorf("DATABASE_URL is not set (set it in .env or environment variables)")
	}

	// Get migrations directory
	migrationsDir := database.GetMigrationsDir(projectDir, project.DBDriver)

	if dryRun {
		return c.dryRun(cmd, dbManager, migrationsDir)
	}

	return c.migrate(cmd, dbManager, migrationsDir, steps)
}

func (c *DBMigrateCommand) dryRun(cmd *cobra.Command, dbManager interfaces.DatabaseManager, migrationsDir string) error {
	r := c.newRenderer(cmd)
	ctx := cmd.Context()
	defer c.flushRenderer(cmd, r)

	// Connect to database
	db, err := dbManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbManager.Close()

	// Create migration runner
	runner, err := database.NewMigrationRunner(db, dbManager.GetDriver(), migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// Get migration status
	statuses, err := runner.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Count pending
	var pending []database.MigrationStatus
	for _, s := range statuses {
		if !s.Applied {
			pending = append(pending, s)
		}
	}

	if len(pending) == 0 {
		r.Title("No pending migrations")
		return nil
	}

	r.Title("Dry run - migrations that would be applied:")
	var body string
	for _, m := range pending {
		body += fmt.Sprintf("  • %s\n", m.Name)
	}
	body += fmt.Sprintf("\n%d pending migration(s).", len(pending))
	r.Section(interfaces.Section{Body: body})

	return nil
}

func (c *DBMigrateCommand) migrate(cmd *cobra.Command, dbManager interfaces.DatabaseManager, migrationsDir string, steps int) error {
	r := c.newRenderer(cmd)
	ctx := cmd.Context()
	defer c.flushRenderer(cmd, r)

	// Connect to database
	db, err := dbManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbManager.Close()

	// Create migration runner
	runner, err := database.NewMigrationRunner(db, dbManager.GetDriver(), migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	r.Title("Running migrations...")

	// Run migrations
	result, err := runner.Up(ctx, steps)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if len(result.Applied) == 0 {
		r.Section(interfaces.Section{Body: "No pending migrations"})
		return nil
	}

	var body string
	for _, m := range result.Applied {
		body += fmt.Sprintf("  ✓ %s\n", m.Name)
	}
	body += fmt.Sprintf("\nSuccessfully applied %d migration(s).", len(result.Applied))
	r.Section(interfaces.Section{Body: body})

	return nil
}
