package templui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/rs/zerolog"
)

type executor struct{}

// NewExecutor creates a new UIExecutor implementation.
func NewExecutor() interfaces.UIExecutor {
	return &executor{}
}

func (e *executor) Version(ctx context.Context, projectDir string) (string, error) {
	logger := zerolog.Ctx(ctx)

	cmd := exec.CommandContext(ctx, "go", "tool", "templui", "--version")
	cmd.Dir = projectDir

	output, err := cmd.Output()
	if err != nil {
		logger.Error().
			Err(err).
			Str("command", "go tool templui --version").
			Str("dir", projectDir).
			Msg("failed to get templui version")
		return "", fmt.Errorf("failed to get templui version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (e *executor) Add(ctx context.Context, projectDir, ref string, components []string, force bool) error {
	logger := zerolog.Ctx(ctx)

	if len(components) == 0 {
		return fmt.Errorf("at least one component name required")
	}

	toolName := "templui"
	if ref != "" {
		toolName = fmt.Sprintf("templui@%s", ref)
	}

	args := []string{"tool", toolName, "add"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, components...)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error().
			Err(err).
			Str("command", fmt.Sprintf("go %s", strings.Join(args, " "))).
			Str("output", string(output)).
			Str("dir", projectDir).
			Msg("failed to add components")
		return fmt.Errorf("failed to add components: %w", err)
	}

	return nil
}

func (e *executor) List(ctx context.Context, projectDir, ref string) ([]interfaces.UIComponent, error) {
	logger := zerolog.Ctx(ctx)

	toolName := "templui"
	if ref != "" {
		toolName = fmt.Sprintf("templui@%s", ref)
	}

	cmd := exec.CommandContext(ctx, "go", "tool", toolName, "list")
	cmd.Dir = projectDir

	output, err := cmd.Output()
	if err != nil {
		logger.Error().
			Err(err).
			Str("command", fmt.Sprintf("go tool %s list", toolName)).
			Str("dir", projectDir).
			Msg("failed to list components")
		return nil, fmt.Errorf("failed to list components: %w", err)
	}

	return parseComponentList(string(output)), nil
}

func (e *executor) Upgrade(ctx context.Context, projectDir, ref string) error {
	logger := zerolog.Ctx(ctx)

	toolName := "templui"
	if ref != "" {
		toolName = fmt.Sprintf("templui@%s", ref)
	}

	cmd := exec.CommandContext(ctx, "go", "tool", toolName, "upgrade")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error().
			Err(err).
			Str("command", fmt.Sprintf("go tool %s upgrade", toolName)).
			Str("output", string(output)).
			Str("dir", projectDir).
			Msg("failed to upgrade templui")
		return fmt.Errorf("failed to upgrade templui: %w", err)
	}

	return nil
}

func (e *executor) IsAvailable(ctx context.Context, projectDir string) bool {
	cmd := exec.CommandContext(ctx, "go", "tool", "templui", "--version")
	cmd.Dir = projectDir
	return cmd.Run() == nil
}

func parseComponentList(output string) []interfaces.UIComponent {
	var components []interfaces.UIComponent

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		components = append(components, interfaces.UIComponent{
			Name: line,
		})
	}

	return components
}
