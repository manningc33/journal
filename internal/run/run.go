// Package run abstracts execution of external programs (the editor, the
// markdown linter and git) behind a small interface so the command layer can be
// unit-tested without spawning real processes.
package run

import (
	"context"
	"os"
	"os/exec"
)

// Runner executes external commands.
type Runner interface {
	// Interactive runs a command attached to the current process's stdio,
	// used for the editor which needs the terminal.
	Interactive(ctx context.Context, name string, args ...string) error
	// Capture runs a command and returns its combined stdout/stderr.
	Capture(ctx context.Context, name string, args ...string) ([]byte, error)
	// LookPath reports the resolved path of name, or an error if not found.
	LookPath(name string) (string, error)
}

// OS is the production Runner backed by os/exec.
type OS struct{}

// Interactive implements Runner.
func (OS) Interactive(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Capture implements Runner.
func (OS) Capture(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// LookPath implements Runner.
func (OS) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
