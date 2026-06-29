// Package editor launches the configured editor on a journal file.
package editor

import (
	"context"
	"strconv"
	"strings"

	"github.com/manningc33/journal/internal/config"
	"github.com/manningc33/journal/internal/run"
)

// Open launches the editor on file. The {line} placeholder is replaced with the
// configured cursor line and {file} with the file path; if no argument contains
// {file}, the path is appended as the final argument.
func Open(ctx context.Context, r run.Runner, cfg config.EditorConfig, file string) error {
	args := make([]string, 0, len(cfg.Args)+1)
	hasFile := false
	for _, a := range cfg.Args {
		a = strings.ReplaceAll(a, "{line}", strconv.Itoa(cfg.Line))
		if strings.Contains(a, "{file}") {
			a = strings.ReplaceAll(a, "{file}", file)
			hasFile = true
		}
		args = append(args, a)
	}
	if !hasFile {
		args = append(args, file)
	}
	return r.Interactive(ctx, cfg.Cmd, args...)
}
