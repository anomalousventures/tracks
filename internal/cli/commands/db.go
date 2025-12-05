package commands

import (
	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

// DBCommand represents the 'db' parent command for database management.
type DBCommand struct {
	detector      interfaces.ProjectDetector
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
}

// NewDBCommand creates a new instance of the 'db' command with injected dependencies.
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

// Command returns the cobra.Command for the 'db' subcommand.
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

	return cmd
}

func (c *DBCommand) run(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}
