package commands

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/internal/database"
	"github.com/spf13/cobra"
)

type DBResetCommand struct {
	detector      interfaces.ProjectDetector
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
	newDBManager  DatabaseManagerFactory
}

func NewDBResetCommand(
	detector interfaces.ProjectDetector,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *DBResetCommand {
	return &DBResetCommand{
		detector:      detector,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
		newDBManager:  DefaultDatabaseManagerFactory(),
	}
}

func NewDBResetCommandWithFactory(
	detector interfaces.ProjectDetector,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
	newDBManager DatabaseManagerFactory,
) *DBResetCommand {
	return &DBResetCommand{
		detector:      detector,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
		newDBManager:  newDBManager,
	}
}

func (c *DBResetCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the database (drop all tables and re-run migrations)",
		Long: `Reset the database by dropping all tables and re-running all migrations.

WARNING: This will delete all data in the database!

Prompts for confirmation unless --force is specified.

Note: This command only supports Postgres projects directly.
For SQLite/go-libsql projects, use: make migrate-reset`,
		Example: `  # Reset database (prompts for confirmation)
  tracks db reset

  # Reset database without confirmation
  tracks db reset --force`,
		RunE: c.runE,
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return cmd
}

func (c *DBResetCommand) runE(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	force, _ := cmd.Flags().GetBool("force")

	project, projectDir, err := c.detector.Detect(ctx, ".")
	if err != nil {
		return fmt.Errorf("not in a Tracks project directory (missing .tracks.yaml): %w", err)
	}

	if project.DBDriver != "postgres" {
		return fmt.Errorf("tracks db reset only supports Postgres projects (found: %s). For SQLite/go-libsql projects, use: make migrate-reset", project.DBDriver)
	}

	dbManager := c.newDBManager(project.DBDriver)
	if err := dbManager.LoadEnv(ctx, projectDir); err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	dbURL := dbManager.GetDatabaseURL()
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is not set (set it in .env or environment variables)")
	}

	if !force {
		confirmed, err := c.promptForConfirmation(cmd, dbURL)
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if !confirmed {
			fmt.Fprintln(cmd.OutOrStdout(), "Reset cancelled.")
			return nil
		}
	}

	r := c.newRenderer(cmd)
	defer c.flushRenderer(cmd, r)

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

	r.Title("Resetting database...")

	result, err := runner.Reset(ctx)
	if err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	var body string
	if len(result.Applied) == 0 {
		body = "No migrations to apply after reset."
	} else {
		body = "Applied migrations:\n"
		for _, m := range result.Applied {
			body += fmt.Sprintf("  ✓ %s\n", m.Name)
		}
		body += fmt.Sprintf("\nReset complete. Applied %d migration(s).", len(result.Applied))
	}
	r.Section(interfaces.Section{Body: body})

	return nil
}

func (c *DBResetCommand) promptForConfirmation(cmd *cobra.Command, dbURL string) (bool, error) {
	warning := `
⚠️  WARNING: This will delete all data in the database!

Database: %s

This action will:
  - Drop all tables
  - Re-run all migrations
  - Delete all existing data

Are you sure? (y/N): `

	fmt.Fprintf(cmd.OutOrStdout(), warning, dbURL)

	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}
