package commands

import (
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/internal/database"
	"github.com/spf13/cobra"
)

type DBRollbackCommand struct {
	detector      interfaces.ProjectDetector
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
}

func NewDBRollbackCommand(
	detector interfaces.ProjectDetector,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *DBRollbackCommand {
	return &DBRollbackCommand{
		detector:      detector,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
	}
}

func (c *DBRollbackCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Roll back database migrations",
		Long: `Roll back database migrations for your Tracks project.

Rolls back the last applied migration by default, or multiple with --steps.

Note: This command only supports Postgres projects directly.
For SQLite/go-libsql projects, use: make migrate-down`,
		Example: `  # Roll back the last migration
  tracks db rollback

  # Roll back 3 migrations
  tracks db rollback --steps 3`,
		RunE: c.runE,
	}

	cmd.Flags().IntP("steps", "n", 1, "Number of migrations to roll back")

	return cmd
}

func (c *DBRollbackCommand) runE(cmd *cobra.Command, _ []string) error {
	r := c.newRenderer(cmd)
	ctx := cmd.Context()
	defer c.flushRenderer(cmd, r)

	steps, _ := cmd.Flags().GetInt("steps")

	// Detect project
	project, projectDir, err := c.detector.Detect(ctx, ".")
	if err != nil {
		return fmt.Errorf("not in a Tracks project directory (missing .tracks.yaml): %w", err)
	}

	// Check for supported driver
	if project.DBDriver != "postgres" {
		return fmt.Errorf("tracks db rollback only supports Postgres projects (found: %s). For SQLite/go-libsql projects, use: make migrate-down", project.DBDriver)
	}

	// Create database manager
	dbManager := database.NewManager(project.DBDriver)
	if err := dbManager.LoadEnv(ctx, projectDir); err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	// Check for DATABASE_URL
	if dbManager.GetDatabaseURL() == "" {
		return fmt.Errorf("DATABASE_URL is not set (set it in .env or environment variables)")
	}

	// Connect to database
	db, err := dbManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbManager.Close()

	// Get migrations directory
	migrationsDir := database.GetMigrationsDir(projectDir, project.DBDriver)

	// Create migration runner
	runner, err := database.NewMigrationRunner(db, dbManager.GetDriver(), migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	r.Title("Rolling back migrations...")

	// Run rollback
	result, err := runner.Down(ctx, steps)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if len(result.Applied) == 0 {
		r.Section(interfaces.Section{Body: "No migrations to roll back"})
		return nil
	}

	var body string
	for _, m := range result.Applied {
		body += fmt.Sprintf("  ✓ %s (rolled back)\n", m.Name)
	}
	body += fmt.Sprintf("\nSuccessfully rolled back %d migration(s).", len(result.Applied))
	r.Section(interfaces.Section{Body: body})

	return nil
}
