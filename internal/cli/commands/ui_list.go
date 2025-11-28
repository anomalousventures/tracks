package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

// UIListCommand represents the 'ui list' subcommand for listing templUI components.
type UIListCommand struct {
	detector      interfaces.ProjectDetector
	executor      interfaces.UIExecutor
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
}

// NewUIListCommand creates a new instance of the 'ui list' command with injected dependencies.
func NewUIListCommand(
	detector interfaces.ProjectDetector,
	executor interfaces.UIExecutor,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *UIListCommand {
	return &UIListCommand{
		detector:      detector,
		executor:      executor,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
	}
}

// Command returns the cobra.Command for the 'ui list' subcommand.
func (c *UIListCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list[@<ref>]",
		Short: "List available templUI components",
		Long: `List all available templUI components and show which ones are installed.

Components that are already installed in your project are marked with ✓.
Use the optional @ref syntax to list components from a specific templUI version.`,
		Example: `  # List available components
  tracks ui list

  # List components from a specific version
  tracks ui list@v0.1.0`,
		RunE: c.runE,
	}

	return cmd
}

func (c *UIListCommand) runE(cmd *cobra.Command, args []string) error {
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

	components, err := c.executor.List(ctx, projectDir, ref)
	if err != nil {
		r.Section(interfaces.Section{
			Body: fmt.Sprintf("Error: %v", err),
		})
		return nil
	}

	installedComponents := getInstalledComponents(projectDir)

	for i := range components {
		if installedComponents[components[i].Name] {
			components[i].Installed = true
		}
	}

	r.Title("Available templUI Components")

	headers := []string{"NAME", "CATEGORY", "INSTALLED"}
	rows := make([][]string, len(components))
	for i, comp := range components {
		installed := "-"
		if comp.Installed {
			installed = "✓"
		}
		category := comp.Category
		if category == "" {
			category = "-"
		}
		rows[i] = []string{comp.Name, category, installed}
	}

	r.Table(interfaces.Table{
		Headers: headers,
		Rows:    rows,
	})

	return nil
}

func getInstalledComponents(projectDir string) map[string]bool {
	uiDir := filepath.Join(projectDir, "internal", "http", "views", "components", "ui")
	installed := make(map[string]bool)

	entries, err := os.ReadDir(uiDir)
	if err != nil {
		return installed
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".templ") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".templ")
		installed[name] = true
	}

	return installed
}
