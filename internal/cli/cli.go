// Package cli wires the internal packages together into the `journal`
// subcommands. It owns argument parsing and orchestration only; all real logic
// lives in the focused packages it calls.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manningc33/journal/internal/config"
	"github.com/manningc33/journal/internal/datefmt"
	"github.com/manningc33/journal/internal/editor"
	"github.com/manningc33/journal/internal/entry"
	"github.com/manningc33/journal/internal/linter"
	"github.com/manningc33/journal/internal/run"
	"github.com/manningc33/journal/internal/vcs"
)

// Version is the build version, overridable via -ldflags at build time.
var Version = "dev"

// App holds the dependencies for a single invocation.
type App struct {
	Out        io.Writer
	Err        io.Writer
	Runner     run.Runner
	ConfigPath string // empty uses config.DefaultPath
}

var offsetRe = regexp.MustCompile(`^-(\d+)$`)

// Run dispatches args[0] to a subcommand.
func (a App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		a.usage()
		return nil
	}
	switch args[0] {
	case "today":
		return a.today(ctx, args[1:])
	case "commit":
		return a.commit(ctx, args[1:])
	case "help", "-h", "--help":
		a.usage()
		return nil
	case "version", "--version":
		fmt.Fprintln(a.Out, Version)
		return nil
	default:
		return fmt.Errorf("unknown command %q (run `journal help`)", args[0])
	}
}

// today opens (creating if needed) the entry for today, or N days ago when
// invoked as `today -N`.
func (a App) today(ctx context.Context, args []string) error {
	days := 0
	forceNew := false
	for _, arg := range args {
		switch {
		case arg == "-h" || arg == "--help":
			fmt.Fprintln(a.Out, "usage: journal today [-N] [-n]\n\nOpens today's entry, creating it if needed.\n  -N   open the entry from N days ago (e.g. -1 = yesterday)\n  -n   create an additional new entry for the day (28-sun2.md, 28-sun3.md, ...)")
			return nil
		case arg == "-n" || arg == "--new":
			forceNew = true
		case offsetRe.MatchString(arg):
			days, _ = strconv.Atoi(offsetRe.FindStringSubmatch(arg)[1])
		default:
			return fmt.Errorf("unexpected argument %q (run `journal today -h`)", arg)
		}
	}

	cfg, err := a.config()
	if err != nil {
		return err
	}

	date := time.Now().AddDate(0, 0, -days)
	e := entry.For(cfg.JournalDir, date, cfg.Format)
	if err := os.MkdirAll(e.Dir, 0o755); err != nil {
		return err
	}

	target := e.Path
	if forceNew {
		target = nextVariant(e.Dir, filepath.Base(e.Path))
	}
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(target, []byte(e.NewContents(cfg.Format)), 0o644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return editor.Open(ctx, a.Runner, cfg.Editor, target)
}

// commit appends a message to an existing entry, lints it and records a git
// commit whose date is derived from the entry's path.
func (a App) commit(ctx context.Context, args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(a.Out, "usage: journal commit <file.md> <message>\n\nAppends <message> to the entry, lints it, and commits it with the date\ntaken from the file's path. All same-day entries (the canonical file plus\nany `today -n` variants) are staged in the commit.")
		return nil
	}
	if len(args) < 2 {
		return errors.New("usage: journal commit <file.md> <message>")
	}
	file := args[0]
	message := strings.TrimSpace(strings.Join(args[1:], " "))
	if message == "" {
		return errors.New("commit message is empty")
	}

	cfg, err := a.config()
	if err != nil {
		return err
	}
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("file %s: %w", file, err)
	}

	// Derive (and validate) the date before mutating anything.
	date, err := entry.DateFromPath(cfg.JournalDir, file, cfg.Format)
	if err != nil {
		return err
	}

	// Append the message to the file being committed, then format and stage
	// every entry for that day (the canonical one plus any `today -n` variants).
	appendText := strings.ReplaceAll(cfg.Commit.AppendFormat, "{message}", message)
	if err := appendToFile(file, appendText); err != nil {
		return err
	}

	canonical := entry.For(cfg.JournalDir, date, cfg.Format)
	files, err := dayFiles(canonical.Dir, filepath.Base(canonical.Path))
	if err != nil {
		return err
	}
	for _, fp := range files {
		if err := linter.Lint(ctx, a.Out, a.Runner, cfg.Linter, fp); err != nil {
			return err
		}
	}

	commitMsg := datefmt.Format(cfg.Format.Commit, date) + " " + message
	if err := vcs.Commit(ctx, a.Runner, cfg.JournalDir, commitMsg, files...); err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "✓ committed %d file(s): %s\n", len(files), commitMsg)
	return nil
}

// nextVariant returns the first same-day entry path that does not yet exist,
// starting at the canonical name then 28-sun2.md, 28-sun3.md, ...
func nextVariant(dir, canonicalBase string) string {
	for n := 1; ; n++ {
		p := filepath.Join(dir, entry.VariantName(canonicalBase, n))
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			return p
		}
	}
}

// dayFiles returns all existing entries for a day (canonical + numeric
// variants) under dir, ordered by variant number.
func dayFiles(dir, canonicalBase string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	type variant struct {
		n int
		p string
	}
	var found []variant
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if n, ok := entry.VariantNum(e.Name(), canonicalBase); ok {
			found = append(found, variant{n, filepath.Join(dir, e.Name())})
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].n < found[j].n })
	files := make([]string, len(found))
	for i, v := range found {
		files[i] = v.p
	}
	return files, nil
}

func (a App) config() (config.Config, error) {
	return config.Load(a.ConfigPath)
}

func (a App) usage() {
	fmt.Fprintln(a.Out, `journal - a terminal journaling tool

usage:
  journal today [-N] [-n]       open today's entry (or N days ago); -n adds a new same-day entry
  journal commit <file> <msg>   append, lint and git-commit an entry (stages all same-day entries)
  journal version               print the version
  journal help                  show this help`)
}

func appendToFile(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		return err
	}
	return f.Close()
}
