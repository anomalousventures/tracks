package generator

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
)

func InitializeGit(ctx context.Context, projectPath string, skipGit bool) error {
	if skipGit {
		return nil
	}

	if err := runGitCommand(ctx, projectPath, "init"); err != nil {
		return errors.New("failed to initialize git repository")
	}

	if err := runGitCommand(ctx, projectPath, "config", "--local", "user.name", "Tracks"); err != nil {
		return errors.New("failed to configure git user")
	}

	if err := runGitCommand(ctx, projectPath, "config", "--local", "user.email", "tracks@tracks.local"); err != nil {
		return errors.New("failed to configure git user")
	}

	if err := runGitCommand(ctx, projectPath, "add", "."); err != nil {
		return errors.New("failed to stage files")
	}

	if err := runGitCommand(ctx, projectPath, "commit", "-m", "Initial commit from Tracks"); err != nil {
		return errors.New("failed to create initial commit")
	}

	return nil
}

func runGitCommand(ctx context.Context, projectPath string, args ...string) error {
	logger := zerolog.Ctx(ctx)
	cmd := exec.Command("git", args...)
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error().
			Err(err).
			Str("command", fmt.Sprintf("git %s", strings.Join(args, " "))).
			Str("output", string(output)).
			Str("dir", projectPath).
			Msg("git command failed")
		return err
	}

	return nil
}
