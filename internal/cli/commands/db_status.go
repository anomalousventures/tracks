package commands

import (
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/internal/database"
	"github.com/spf13/cobra"
)

type DBStatusCommand struct {
	detector      interfaces.ProjectDetector
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
	newDBManager  DatabaseManagerFactory
}

func NewDBStatusCommand(
	detector interfaces.ProjectDetector,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *DBStatusCommand {
	return &DBStatusCommand{
		detector:      detector,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
		newDBManager:  DefaultDatabaseManagerFactory(),
	}
}

func NewDBStatusCommandWithFactory(
	detector interfaces.ProjectDetector,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
	newDBManager DatabaseManagerFactory,
) *DBStatusCommand {
	return &DBStatusCommand{
		detector:      detector,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
		newDBManager:  newDBManager,
	}
}

func (c *DBStatusCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show database migration status",
		Long: `Show the status of database migrations for your Tracks project.

Displays which migrations have been applied and which are pending.

Note: This command only supports Postgres projects directly.
For SQLite/go-libsql projects, use: make migrate-status`,
		Example: `  # Show migration status
  tracks db status`,
		RunE: c.runE,
	}

	return cmd
}

func (c *DBStatusCommand) runE(cmd *cobra.Command, _ []string) error {
	r := c.newRenderer(cmd)
	ctx := cmd.Context()
	defer c.flushRenderer(cmd, r)

	project, projectDir, err := c.detector.Detect(ctx, ".")
	if err != nil {
		return fmt.Errorf("not in a Tracks project directory (missing .tracks.yaml): %w", err)
	}

	if project.DBDriver != "postgres" {
		return fmt.Errorf("tracks db status only supports Postgres projects (found: %s). For SQLite/go-libsql projects, use: make migrate-status", project.DBDriver)
	}

	dbManager := c.newDBManager(project.DBDriver)
	if err := dbManager.LoadEnv(ctx, projectDir); err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	dbURL := dbManager.GetDatabaseURL()
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is not set (set it in .env or environment variables)")
	}

	migrationsDir := database.GetMigrationsDir(projectDir, project.DBDriver)

	db, err := dbManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbManager.Close()

	runner, err := database.NewMigrationRunner(db, dbManager.GetDriver(), migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	statuses, err := runner.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	var applied, pending []database.MigrationStatus
	for _, s := range statuses {
		if s.Applied {
			applied = append(applied, s)
		} else {
			pending = append(pending, s)
		}
	}

	var body string
	body += fmt.Sprintf("Database: %s\n", database.SanitizeURL(dbURL))
	body += fmt.Sprintf("Driver: %s\n\n", project.DBDriver)

	if len(applied) > 0 {
		body += "Applied Migrations:\n"
		for _, m := range applied {
			body += fmt.Sprintf("  ✓ %s\n", m.Name)
		}
		body += "\n"
	}

	if len(pending) > 0 {
		body += "Pending Migrations:\n"
		for _, m := range pending {
			body += fmt.Sprintf("  ○ %s\n", m.Name)
		}
		body += "\n"
	}

	body += fmt.Sprintf("Total: %d applied, %d pending", len(applied), len(pending))

	r.Section(interfaces.Section{Body: body})

	return nil
}
