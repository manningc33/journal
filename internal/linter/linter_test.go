package linter

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/manningc33/journal/internal/config"
	"github.com/manningc33/journal/internal/run"
)

func TestLintDisabledIsNoop(t *testing.T) {
	fake := &run.Fake{}
	cfg := config.LinterConfig{Enabled: false}
	if err := Lint(context.Background(), io.Discard, fake, cfg, "x.md"); err != nil {
		t.Fatal(err)
	}
	if len(fake.Calls) != 0 {
		t.Errorf("expected no calls, got %v", fake.Calls)
	}
}

func TestLintRunsWhenFound(t *testing.T) {
	fake := &run.Fake{}
	cfg := config.LinterConfig{Enabled: true, Cmd: "markdownlint-cli2", Args: []string{"--fix"}}
	if err := Lint(context.Background(), io.Discard, fake, cfg, "x.md"); err != nil {
		t.Fatal(err)
	}
	if len(fake.Calls) != 1 {
		t.Fatalf("calls = %v", fake.Calls)
	}
	got := fake.Calls[0]
	if got.Kind != "capture" || got.Args[len(got.Args)-1] != "x.md" {
		t.Errorf("call = %+v", got)
	}
}

func TestLintSkipsWhenMissing(t *testing.T) {
	fake := &run.Fake{NotFound: map[string]bool{"markdownlint-cli2": true}}
	cfg := config.LinterConfig{Enabled: true, Cmd: "markdownlint-cli2", FallbackPath: "/nonexistent/bin"}
	var buf strings.Builder
	if err := Lint(context.Background(), &buf, fake, cfg, "x.md"); err != nil {
		t.Fatal(err)
	}
	if len(fake.Calls) != 0 {
		t.Errorf("expected lint skipped, got calls %v", fake.Calls)
	}
	if !strings.Contains(buf.String(), "skipping lint") {
		t.Errorf("expected warning, got %q", buf.String())
	}
}
