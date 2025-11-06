package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	stepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14"))

	checkmark = successStyle.Render("✓")
)

type SuccessOutput struct {
	ProjectName    string
	ProjectPath    string
	ModulePath     string
	DatabaseDriver string
	GitInitialized bool
	NoColor        bool
}

func RenderSuccessOutput(output SuccessOutput) string {
	if output.NoColor {
		return renderSuccessPlain(output)
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(successStyle.Render(fmt.Sprintf("%s Project '%s' created successfully!", checkmark, output.ProjectName)))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Location:      "))
	b.WriteString(valueStyle.Render(output.ProjectPath))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Module:        "))
	b.WriteString(valueStyle.Render(output.ModulePath))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Database:      "))
	b.WriteString(valueStyle.Render(output.DatabaseDriver))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Git:           "))
	if output.GitInitialized {
		b.WriteString(successStyle.Render("initialized"))
	} else {
		b.WriteString(valueStyle.Render("not initialized"))
	}
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Next steps:"))
	b.WriteString("\n")

	steps := []string{
		fmt.Sprintf("cd %s", output.ProjectName),
		"go mod download",
		"make test",
		"make dev",
	}

	for i, step := range steps {
		b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, stepStyle.Render(step)))
	}

	return b.String()
}

func renderSuccessPlain(output SuccessOutput) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("✓ Project '%s' created successfully!", output.ProjectName))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Location:      %s\n", output.ProjectPath))
	b.WriteString(fmt.Sprintf("Module:        %s\n", output.ModulePath))
	b.WriteString(fmt.Sprintf("Database:      %s\n", output.DatabaseDriver))

	gitStatus := "not initialized"
	if output.GitInitialized {
		gitStatus = "initialized"
	}
	b.WriteString(fmt.Sprintf("Git:           %s\n\n", gitStatus))

	b.WriteString("Next steps:\n")
	b.WriteString(fmt.Sprintf("  1. cd %s\n", output.ProjectName))
	b.WriteString("  2. go mod download\n")
	b.WriteString("  3. make test\n")
	b.WriteString("  4. make dev\n")

	return b.String()
}

func GetAbsolutePath(basePath, projectName string) (string, error) {
	projectPath := filepath.Join(basePath, projectName)
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return projectPath, err
	}
	return absPath, nil
}
