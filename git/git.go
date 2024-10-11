package git

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
)

func isRepo(ctx context.Context, repo string) bool {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = repo
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func hasStagedChanges(ctx context.Context, repo string) bool {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	cmd.Dir = repo
	if err := cmd.Run(); err != nil {
		return cmd.ProcessState.ExitCode() == 1
	}
	return cmd.ProcessState.ExitCode() == 1
}

// creates a git repository in path if it does not yet exist
func CreateRepoIfNotExists(ctx context.Context, repo, origin string) error {
	if isRepo(ctx, repo) {
		return nil
	}

	cmd := exec.CommandContext(ctx, "git", "clone", origin)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	return nil
}

func CommitAndPushChanges(ctx context.Context, repo, msg string) error {
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = repo
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	if !hasStagedChanges(ctx, repo) {
		slog.Info("no staged changes to commit")
		return nil
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", msg)
	cmd.Dir = repo
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "push", "origin", "master")
	cmd.Dir = repo
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}
	return nil
}

func ConfigureUser(ctx context.Context, name, email string) error {
	cmd := exec.CommandContext(ctx, "git", "config", "--global", "user.name", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed set user name", err)
	}

	cmd = exec.CommandContext(ctx, "git", "config", "--global", "user.email", email)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed set user email", err)
	}
	return nil
}
