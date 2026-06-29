// Package vcs wraps the git operations used to record a journal entry.
package vcs

import (
	"context"
	"fmt"

	"github.com/manningc33/journal/internal/run"
)

// Commit stages files and commits them with message, operating inside repoDir.
func Commit(ctx context.Context, r run.Runner, repoDir, message string, files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to commit")
	}
	addArgs := append([]string{"-C", repoDir, "add"}, files...)
	if out, err := r.Capture(ctx, "git", addArgs...); err != nil {
		return fmt.Errorf("git add: %w: %s", err, out)
	}
	if out, err := r.Capture(ctx, "git", "-C", repoDir, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w: %s", err, out)
	}
	return nil
}
