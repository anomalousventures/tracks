package commands

import (
	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

type DBCommand struct {
	detector      interfaces.ProjectDetector
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
}

func NewDBCommand(
	detector interfaces.ProjectDetector,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *DBCommand {
	return &DBCommand{
		detector:      detector,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
	}
}

func (c *DBCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management commands",
		Long: `Database management commands for your Tracks project.

Commands for managing database migrations, checking migration status,
and resetting the database.

This command must be run from within a Tracks project (containing .tracks.yaml).`,
		Example: `  # Run pending migrations
  tracks db migrate

  # Roll back the last migration
  tracks db rollback

  # Check migration status
  tracks db status

  # Reset the database (drop and recreate)
  tracks db reset`,
		Run: c.run,
	}

	// Add subcommands
	migrateCmd := NewDBMigrateCommand(c.detector, c.newRenderer, c.flushRenderer)
	cmd.AddCommand(migrateCmd.Command())

	rollbackCmd := NewDBRollbackCommand(c.detector, c.newRenderer, c.flushRenderer)
	cmd.AddCommand(rollbackCmd.Command())

	statusCmd := NewDBStatusCommand(c.detector, c.newRenderer, c.flushRenderer)
	cmd.AddCommand(statusCmd.Command())

	resetCmd := NewDBResetCommand(c.detector, c.newRenderer, c.flushRenderer)
	cmd.AddCommand(resetCmd.Command())

	return cmd
}

func (c *DBCommand) run(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}
