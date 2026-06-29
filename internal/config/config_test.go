package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Editor.Cmd != "nvim" || cfg.Linter.Cmd != "markdownlint-cli2" {
		t.Errorf("defaults not applied: %+v", cfg)
	}
	if cfg.Format.Dir != "%Y/%m-%b" {
		t.Errorf("default dir format = %q", cfg.Format.Dir)
	}
}

func TestLoadOverlaysFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
journal_dir = "` + dir + `"

[editor]
cmd = "vim"

[linter]
enabled = false
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Editor.Cmd != "vim" {
		t.Errorf("editor.cmd = %q, want vim", cfg.Editor.Cmd)
	}
	if cfg.Linter.Enabled {
		t.Error("linter.enabled should be false")
	}
	// Unspecified fields keep their defaults.
	if cfg.Editor.Line != 3 || cfg.Format.Commit != "%Y-%m-%d" {
		t.Errorf("defaults clobbered: %+v", cfg)
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := expandPath("~/journal"); got != filepath.Join(home, "journal") {
		t.Errorf("expandPath(~/journal) = %q", got)
	}
	t.Setenv("FOO", "bar")
	if got := expandPath("$FOO/x"); got != "bar/x" {
		t.Errorf("expandPath($FOO/x) = %q", got)
	}
}

func TestRenderHeaderMatchesExample(t *testing.T) {
	f := Defaults().Format
	d := time.Date(2026, 6, 28, 16, 43, 0, 0, time.UTC)
	if got := f.Render(f.Header, d); got != "# june 28, 2026 (sunday 16:43)" {
		t.Errorf("header = %q", got)
	}
}

func TestRenderSqueezesSinglDigitDay(t *testing.T) {
	f := Defaults().Format
	d := time.Date(2026, 6, 7, 9, 5, 0, 0, time.UTC) // single-digit day -> %e pads with a space
	if got := f.Render(f.Header, d); got != "# june 7, 2026 (sunday 09:05)" {
		t.Errorf("header = %q (spaces not squeezed?)", got)
	}
}
