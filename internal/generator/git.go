package generator

import (
	"fmt"
	"os/exec"
)

func InitializeGit(projectPath string, skipGit bool) error {
	if skipGit {
		return nil
	}

	if err := runGitCommand(projectPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	if err := runGitCommand(projectPath, "add", "."); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	if err := runGitCommand(projectPath, "commit", "-m", "Initial commit from Tracks"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

func runGitCommand(projectPath string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = projectPath

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
