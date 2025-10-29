package commands

import (
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

// BuildInfo provides version metadata for the CLI.
type BuildInfo interface {
	GetVersion() string
	GetCommit() string
	GetDate() string
}

// VersionCommand represents the 'version' command for displaying version information.
type VersionCommand struct {
	build         BuildInfo
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
}

// NewVersionCommand creates a new instance of the 'version' command with injected dependencies.
func NewVersionCommand(build BuildInfo, newRenderer RendererFactory, flushRenderer RendererFlusher) *VersionCommand {
	return &VersionCommand{
		build:         build,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
	}
}

// Command returns the cobra.Command for the 'version' subcommand.
func (c *VersionCommand) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Display the version number, git commit hash, and build date for this Tracks CLI binary.",
		Run:   c.run,
	}
}

func (c *VersionCommand) run(cmd *cobra.Command, args []string) {
	r := c.newRenderer(cmd)

	r.Title(fmt.Sprintf("Tracks %s", c.build.GetVersion()))
	r.Section(interfaces.Section{
		Body: fmt.Sprintf("Commit: %s\nBuilt: %s", c.build.GetCommit(), c.build.GetDate()),
	})

	c.flushRenderer(cmd, r)
}
