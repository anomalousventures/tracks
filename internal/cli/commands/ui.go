package commands

import (
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

// UICommand represents the 'ui' parent command for UI component management.
type UICommand struct {
	detector      interfaces.ProjectDetector
	executor      interfaces.UIExecutor
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
}

// NewUICommand creates a new instance of the 'ui' command with injected dependencies.
func NewUICommand(
	detector interfaces.ProjectDetector,
	executor interfaces.UIExecutor,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *UICommand {
	return &UICommand{
		detector:      detector,
		executor:      executor,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
	}
}

// Command returns the cobra.Command for the 'ui' subcommand.
func (c *UICommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Manage templUI components in your Tracks project",
		Long: `Manage templUI components in your Tracks project.

templUI is a component library for Go templ that follows the shadcn/ui
philosophy: copy code, own components. Components are copied to your
project's internal/http/views/components/ui/ directory.

This command must be run from within a Tracks project (containing .tracks.yaml).`,
		Example: `  # Show templUI version
  tracks ui --version

  # Add a button component
  tracks ui add button

  # Add multiple components
  tracks ui add button card toast

  # List available components
  tracks ui list

  # Upgrade templUI to latest
  tracks ui upgrade`,
		Run: c.run,
	}

	cmd.Flags().Bool("version", false, "Show templUI version")

	addCmd := NewUIAddCommand(c.detector, c.executor, c.newRenderer, c.flushRenderer)
	cmd.AddCommand(addCmd.Command())

	listCmd := NewUIListCommand(c.detector, c.executor, c.newRenderer, c.flushRenderer)
	cmd.AddCommand(listCmd.Command())

	upgradeCmd := NewUIUpgradeCommand(c.detector, c.executor, c.newRenderer, c.flushRenderer)
	cmd.AddCommand(upgradeCmd.Command())

	return cmd
}

func (c *UICommand) run(cmd *cobra.Command, args []string) {
	r := c.newRenderer(cmd)
	ctx := cmd.Context()

	showVersion, _ := cmd.Flags().GetBool("version")
	if showVersion {
		version, err := c.executor.Version(ctx, ".")
		if err != nil {
			r.Section(interfaces.Section{
				Body: fmt.Sprintf("Error: %v", err),
			})
			c.flushRenderer(cmd, r)
			return
		}
		r.Title(fmt.Sprintf("templUI %s", version))
		c.flushRenderer(cmd, r)
		return
	}

	_ = cmd.Help()
}
