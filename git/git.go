package git

import (
	"context"
	"fmt"
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
