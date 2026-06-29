// Package linter runs the configured markdown formatter over an entry. The step
// is best-effort: if disabled or the binary cannot be found it is skipped with a
// warning rather than failing the commit, matching the original script.
package linter

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/manningc33/journal/internal/config"
	"github.com/manningc33/journal/internal/run"
)

// Lint formats file in place. Warnings are written to w; a hard failure of the
// linter itself (when present) is returned as an error.
func Lint(ctx context.Context, w io.Writer, r run.Runner, cfg config.LinterConfig, file string) error {
	if !cfg.Enabled {
		return nil
	}

	bin, ok := resolve(r, cfg)
	if !ok {
		fmt.Fprintf(w, "! warning: %s not found; skipping lint\n", cfg.Cmd)
		return nil
	}

	args := append(append([]string{}, cfg.Args...), file)
	if out, err := r.Capture(ctx, bin, args...); err != nil {
		return fmt.Errorf("lint %s: %w: %s", file, err, out)
	}
	fmt.Fprintf(w, "✓ formatted with %s\n", cfg.Cmd)
	return nil
}

// resolve finds the linter binary: first on PATH, then the configured fallback.
func resolve(r run.Runner, cfg config.LinterConfig) (string, bool) {
	if path, err := r.LookPath(cfg.Cmd); err == nil {
		return path, true
	}
	if cfg.FallbackPath != "" {
		if _, err := os.Stat(cfg.FallbackPath); err == nil {
			return cfg.FallbackPath, true
		}
	}
	return "", false
}
