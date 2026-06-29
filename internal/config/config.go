// Package config loads journal settings from a TOML file, falling back to
// defaults that reproduce the behavior of the original shell scripts. Every
// user-tunable knob (paths, editor, linter and the date formats) lives here.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/manningc33/journal/internal/datefmt"
)

// Config is the fully resolved journal configuration.
type Config struct {
	// JournalDir is the root under which dated entries are stored.
	JournalDir string       `toml:"journal_dir"`
	Editor     EditorConfig `toml:"editor"`
	Linter     LinterConfig `toml:"linter"`
	Format     FormatConfig `toml:"format"`
	Commit     CommitConfig `toml:"commit"`
}

// EditorConfig controls how an entry is opened.
type EditorConfig struct {
	Cmd string `toml:"cmd"`
	// Args may contain the {line} and {file} placeholders. If {file} is
	// absent the file path is appended as the final argument.
	Args []string `toml:"args"`
	// Line is the cursor line the editor opens on.
	Line int `toml:"line"`
}

// LinterConfig controls the optional markdown lint/format step on commit.
type LinterConfig struct {
	Enabled bool     `toml:"enabled"`
	Cmd     string   `toml:"cmd"`
	Args    []string `toml:"args"`
	// FallbackPath is used when Cmd is not found on PATH.
	FallbackPath string `toml:"fallback_path"`
}

// FormatConfig holds the strftime-style date patterns. These mirror the
// original scripts' intentional formatting and should round-trip exactly.
type FormatConfig struct {
	Dir       string          `toml:"dir"`    // directory under JournalDir, e.g. "%Y/%m-%b"
	File      string          `toml:"file"`   // filename, e.g. "%d-%a.md"
	Header    string          `toml:"header"` // first line of a new entry
	Commit    string          `toml:"commit"` // date prefix of the commit message
	Transform TransformConfig `toml:"transform"`
}

// TransformConfig applies text normalisation after date formatting, matching
// the original scripts' `tr -s ' '` (squeeze) and lowercasing.
type TransformConfig struct {
	Lowercase     bool `toml:"lowercase"`
	SqueezeSpaces bool `toml:"squeeze_spaces"`
}

// CommitConfig controls the text appended to an entry when committed.
type CommitConfig struct {
	// AppendFormat is appended verbatim; {message} is replaced with the
	// commit message.
	AppendFormat string `toml:"append_format"`
}

// Defaults returns a Config equivalent to the original today.sh / jcommit.sh.
func Defaults() Config {
	return Config{
		JournalDir: "~/journal",
		Editor: EditorConfig{
			Cmd:  "nvim",
			Args: []string{"+{line}"},
			Line: 3,
		},
		Linter: LinterConfig{
			Enabled:      true,
			Cmd:          "markdownlint-cli2",
			Args:         []string{"--fix"},
			FallbackPath: "~/.local/share/nvim/mason/bin/markdownlint-cli2",
		},
		Format: FormatConfig{
			Dir:       "%Y/%m-%b",
			File:      "%d-%a.md",
			Header:    "# %B %e, %Y (%A %H:%M)",
			Commit:    "%Y-%m-%d",
			Transform: TransformConfig{Lowercase: true, SqueezeSpaces: true},
		},
		Commit: CommitConfig{AppendFormat: "\n\n> {message}\n"},
	}
}

// DefaultPath returns the XDG-aware config location.
func DefaultPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = filepath.Join(home(), ".config")
	}
	return filepath.Join(dir, "journal", "config.toml")
}

// Load reads configuration from path, layering it over Defaults. An empty path
// uses DefaultPath; a missing file is not an error (defaults are used).
func Load(path string) (Config, error) {
	cfg := Defaults()
	if path == "" {
		path = DefaultPath()
	}
	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return cfg, fmt.Errorf("config %s: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return cfg, fmt.Errorf("config %s: %w", path, err)
	}

	cfg.JournalDir = expandPath(cfg.JournalDir)
	cfg.Linter.FallbackPath = expandPath(cfg.Linter.FallbackPath)
	if cfg.JournalDir == "" {
		return cfg, errors.New("journal_dir is not set")
	}
	return cfg, nil
}

// Render formats t with the given strftime pattern and applies the configured
// text transforms.
func (f FormatConfig) Render(pattern string, t time.Time) string {
	return f.Transform.apply(datefmt.Format(pattern, t))
}

var multiSpace = regexp.MustCompile(` +`)

func (t TransformConfig) apply(s string) string {
	if t.SqueezeSpaces {
		s = multiSpace.ReplaceAllString(s, " ")
	}
	if t.Lowercase {
		s = strings.ToLower(s)
	}
	return s
}

// expandPath resolves a leading ~ and any $ENV references.
func expandPath(p string) string {
	if p == "" {
		return p
	}
	p = os.ExpandEnv(p)
	switch {
	case p == "~":
		return home()
	case strings.HasPrefix(p, "~/"):
		return filepath.Join(home(), p[2:])
	}
	return p
}

func home() string {
	h, _ := os.UserHomeDir()
	return h
}
