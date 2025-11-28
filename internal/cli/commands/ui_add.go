package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

const (
	scriptMarkerBegin = "<!-- TRACKS:UI_SCRIPTS:BEGIN -->"
	scriptMarkerEnd   = "<!-- TRACKS:UI_SCRIPTS:END -->"
)

// UIAddCommand represents the 'ui add' subcommand for adding templUI components.
type UIAddCommand struct {
	detector      interfaces.ProjectDetector
	executor      interfaces.UIExecutor
	newRenderer   RendererFactory
	flushRenderer RendererFlusher
}

// NewUIAddCommand creates a new instance of the 'ui add' command with injected dependencies.
func NewUIAddCommand(
	detector interfaces.ProjectDetector,
	executor interfaces.UIExecutor,
	newRenderer RendererFactory,
	flushRenderer RendererFlusher,
) *UIAddCommand {
	return &UIAddCommand{
		detector:      detector,
		executor:      executor,
		newRenderer:   newRenderer,
		flushRenderer: flushRenderer,
	}
}

// Command returns the cobra.Command for the 'ui add' subcommand.
func (c *UIAddCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add[@<ref>] <component> [component...]",
		Short: "Add templUI components to your project",
		Long: `Add one or more templUI components to your Tracks project.

Components are copied to your project's internal/http/views/components/ui/
directory. If a component has a Script() function, it will be automatically
injected into your base.templ layout between the TRACKS:UI_SCRIPTS markers.

The optional @ref syntax allows you to specify a templUI version:
  tracks ui add@v0.1.0 button    # Use specific version
  tracks ui add button           # Use current version`,
		Example: `  # Add a single component
  tracks ui add button

  # Add multiple components
  tracks ui add button card toast

  # Add components from a specific version
  tracks ui add@v0.1.0 button card

  # Force overwrite existing components
  tracks ui add button --force`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.runE,
	}

	cmd.Flags().BoolP("force", "f", false, "Overwrite existing components without prompting")

	return cmd
}

func (c *UIAddCommand) runE(cmd *cobra.Command, args []string) error {
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
	force, _ := cmd.Flags().GetBool("force")

	r.Title("Adding templUI components")

	if err := c.executor.Add(ctx, projectDir, ref, args, force); err != nil {
		r.Section(interfaces.Section{
			Body: fmt.Sprintf("Error: %v", err),
		})
		return nil
	}

	injectedScripts := c.injectScripts(projectDir, args)

	var body strings.Builder
	body.WriteString(fmt.Sprintf("Added %d component(s): %s", len(args), strings.Join(args, ", ")))
	if len(injectedScripts) > 0 {
		body.WriteString(fmt.Sprintf("\nInjected scripts: %s", strings.Join(injectedScripts, ", ")))
	}

	r.Section(interfaces.Section{
		Body: body.String(),
	})

	return nil
}

func (c *UIAddCommand) injectScripts(projectDir string, components []string) []string {
	baseTemplPath := filepath.Join(projectDir, "internal", "http", "views", "layouts", "base.templ")
	uiDir := filepath.Join(projectDir, "internal", "http", "views", "components", "ui")

	baseContent, err := os.ReadFile(baseTemplPath)
	if err != nil {
		return nil
	}

	var injectedScripts []string
	contentStr := string(baseContent)

	for _, component := range components {
		componentFile := filepath.Join(uiDir, component+".templ")
		scriptFuncName := capitalizeFirst(component) + "Script"

		if !hasScriptFunc(componentFile, scriptFuncName) {
			continue
		}

		scriptCall := fmt.Sprintf("@ui.%s()", scriptFuncName)
		if strings.Contains(contentStr, scriptCall) {
			continue
		}

		contentStr = injectScriptCall(contentStr, scriptCall)
		injectedScripts = append(injectedScripts, scriptFuncName)
	}

	if len(injectedScripts) > 0 {
		_ = os.WriteFile(baseTemplPath, []byte(contentStr), 0644)
	}

	return injectedScripts
}

func hasScriptFunc(componentFile, funcName string) bool {
	content, err := os.ReadFile(componentFile)
	if err != nil {
		return false
	}

	pattern := fmt.Sprintf(`(?m)^(templ|func)\s+%s\s*\(`, regexp.QuoteMeta(funcName))
	re := regexp.MustCompile(pattern)
	return re.Match(content)
}

func injectScriptCall(content, scriptCall string) string {
	beginIdx := strings.Index(content, scriptMarkerBegin)
	endIdx := strings.Index(content, scriptMarkerEnd)

	if beginIdx == -1 || endIdx == -1 || endIdx <= beginIdx {
		return content
	}

	insertPos := beginIdx + len(scriptMarkerBegin)

	indentation := "\t\t\t"
	if lineStart := strings.LastIndex(content[:beginIdx], "\n"); lineStart != -1 {
		lineContent := content[lineStart+1 : beginIdx]
		indentation = strings.TrimSuffix(lineContent, strings.TrimLeft(lineContent, " \t"))
	}

	newLine := "\n" + indentation + scriptCall
	return content[:insertPos] + newLine + content[insertPos:]
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func parseRefFromUse(useStr, calledAs string) string {
	if calledAs == useStr || strings.Contains(calledAs, "<") {
		return ""
	}
	if idx := strings.Index(calledAs, "@"); idx != -1 {
		return calledAs[idx+1:]
	}
	return ""
}
