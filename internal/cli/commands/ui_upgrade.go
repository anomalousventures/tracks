package commands

import (
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

// UIUpgradeCommand represents the 'ui upgrade' subcommand for updating templUI.
type UIUpgradeCommand struct {
	detector      interfaces.ProjectDetector
	executor      interfaces.UIExecutor
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
}

// NewUIUpgradeCommand creates a new instance of the 'ui upgrade' command with injected dependencies.
func NewUIUpgradeCommand(
	detector interfaces.ProjectDetector,
	executor interfaces.UIExecutor,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *UIUpgradeCommand {
	return &UIUpgradeCommand{
		detector:      detector,
		executor:      executor,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
	}
}

// Command returns the cobra.Command for the 'ui upgrade' subcommand.
func (c *UIUpgradeCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade[@<ref>]",
		Short: "Upgrade templUI to the latest or specified version",
		Long: `Upgrade the templUI tool to a newer version.

Without a version specified, upgrades to the latest available version.
Use the @ref syntax to upgrade to a specific version.`,
		Example: `  # Upgrade to latest version
  tracks ui upgrade

  # Upgrade to a specific version
  tracks ui upgrade@v0.2.0`,
		RunE: c.runE,
	}

	return cmd
}

func (c *UIUpgradeCommand) runE(cmd *cobra.Command, args []string) error {
	r := c.newRenderer(cmd)
	ctx := cmd.Context()
	defer c.flushRenderer(cmd, r)

	project, projectDir, err := c.detector.Detect(ctx, ".")
	if err != nil {
		r.Section(interfaces.Section{
			Body: fmt.Sprintf("Error: %v", err),
		})
		return nil
	}
	if project == nil {
		r.Section(interfaces.Section{
			Body: "Error: not in a Tracks project (no .tracks.yaml found)",
		})
		return nil
	}

	ref := parseRefFromUse(cmd.Use, cmd.CalledAs())

	r.Title("Upgrading templUI")

	if err := c.executor.Upgrade(ctx, projectDir, ref); err != nil {
		r.Section(interfaces.Section{
			Body: fmt.Sprintf("Error: %v", err),
		})
		return nil
	}

	version, err := c.executor.Version(ctx, projectDir)
	if err != nil {
		r.Section(interfaces.Section{
			Body: "Upgrade completed successfully",
		})
		return nil
	}

	targetMsg := "latest"
	if ref != "" {
		targetMsg = ref
	}

	r.Section(interfaces.Section{
		Body: fmt.Sprintf("Successfully upgraded templUI to %s (version: %s)", targetMsg, version),
	})

	return nil
}
