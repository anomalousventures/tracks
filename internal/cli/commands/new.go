package commands

import (
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

// RendererFactory creates a renderer from a cobra command.
type RendererFactory func(*cobra.Command) interfaces.Renderer

// RendererFlusher flushes a renderer and handles errors.
type RendererFlusher func(*cobra.Command, interfaces.Renderer)

// NewCommand represents the 'new' command for creating Tracks applications.
// Follows ADR-001 dependency injection pattern: command struct with injected dependencies.
type NewCommand struct {
	validator     interfaces.Validator
	generator     interfaces.ProjectGenerator
	newRenderer   RendererFactory
	flushRenderer RendererFlusher

	// Flags
	dbDriver   string
	modulePath string
	noGit      bool
}

// NewNewCommand creates a new instance of the 'new' command with injected dependencies.
// Follows ADR-001: constructor accepts all dependencies as parameters.
func NewNewCommand(validator interfaces.Validator, generator interfaces.ProjectGenerator, newRenderer RendererFactory, flushRenderer RendererFlusher) *NewCommand {
	return &NewCommand{
		validator:     validator,
		generator:     generator,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
	}
}

// Command returns the cobra.Command for the 'new' subcommand.
func (c *NewCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Tracks application",
		Long: `Create a new Tracks application with the specified project name.

This command generates a complete Go web application with:
  - Proper project structure following Go best practices
  - Type-safe templates using templ
  - Type-safe SQL queries using SQLC
  - Built-in authentication and authorization (RBAC)
  - Development tooling (Makefile, hot-reload, linting)
  - Docker and CI/CD configurations

The generated application is production-ready and follows idiomatic Go patterns.`,
		Example: `  # Create a new application with default settings
  tracks new myapp

  # Specify database driver
  tracks new myapp --db postgres

  # Custom Go module path
  tracks new myapp --module github.com/myorg/myapp

  # Skip git initialization
  tracks new myapp --no-git

  # Combine flags
  tracks new myapp --db postgres --module github.com/myorg/myapp --no-git`,
		Args: cobra.ExactArgs(1),
		RunE: c.run,
	}

	// Add flags
	cmd.Flags().StringVar(&c.dbDriver, "db", "go-libsql", "Database driver (go-libsql|sqlite3|postgres)")
	cmd.Flags().StringVar(&c.modulePath, "module", "", "Go module path (e.g., github.com/user/project)")
	cmd.Flags().BoolVar(&c.noGit, "no-git", false, "Skip git repository initialization")

	return cmd
}

func (c *NewCommand) run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	projectName := args[0]

	// Validate project name
	if err := c.validator.ValidateProjectName(ctx, projectName); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	// Validate database driver
	if err := c.validator.ValidateDatabaseDriver(ctx, c.dbDriver); err != nil {
		return fmt.Errorf("invalid database driver: %w", err)
	}

	// Validate or generate module path
	if c.modulePath == "" {
		// Auto-generate from project name
		c.modulePath = fmt.Sprintf("example.com/%s", projectName)
	} else {
		// Validate provided module path
		if err := c.validator.ValidateModulePath(ctx, c.modulePath); err != nil {
			return fmt.Errorf("invalid module path: %w", err)
		}
	}

	r := c.newRenderer(cmd)

	r.Title(fmt.Sprintf("Creating new Tracks application: %s", projectName))
	r.Section(interfaces.Section{
		Body: fmt.Sprintf("Database: %s\nModule: %s\nGit: %t\n\n(Full implementation coming soon)",
			c.dbDriver, c.modulePath, !c.noGit),
	})

	c.flushRenderer(cmd, r)
	return nil
}
