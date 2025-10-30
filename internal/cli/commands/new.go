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
	return &cobra.Command{
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

  # Future: Specify database driver
  tracks new myapp --db postgres

  # Future: Custom Go module path
  tracks new myapp --module github.com/myorg/myapp

  # Future: Skip git initialization
  tracks new myapp --no-git`,
		Args: cobra.ExactArgs(1),
		RunE: c.run,
	}
}

func (c *NewCommand) run(cmd *cobra.Command, args []string) error {
	projectName := args[0]
	r := c.newRenderer(cmd)

	r.Title(fmt.Sprintf("Creating new Tracks application: %s", projectName))
	r.Section(interfaces.Section{
		Body: "(Full implementation coming soon)",
	})

	c.flushRenderer(cmd, r)
	return nil
}
