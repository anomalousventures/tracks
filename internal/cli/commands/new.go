package commands

import (
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/renderer"
	"github.com/spf13/cobra"
)

// RendererFactory creates a renderer from a cobra command.
type RendererFactory func(*cobra.Command) renderer.Renderer

// RendererFlusher flushes a renderer and handles errors.
type RendererFlusher func(*cobra.Command, renderer.Renderer)

// NewCmd creates the 'new' command for creating new Tracks applications.
func NewCmd(newRenderer RendererFactory, flushRenderer RendererFlusher) *cobra.Command {
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
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]
			r := newRenderer(cmd)

			r.Title(fmt.Sprintf("Creating new Tracks application: %s", projectName))
			r.Section(renderer.Section{
				Body: "(Full implementation coming soon)",
			})

			flushRenderer(cmd, r)
		},
	}
}
