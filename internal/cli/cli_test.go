package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/manningc33/journal/internal/run"
)

// newApp writes a config pointing journal_dir at a temp dir and returns a wired
// App plus the journal root and a fake runner for assertions.
func newApp(t *testing.T) (App, string, *run.Fake) {
	t.Helper()
	journalDir := t.TempDir()
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(cfgPath, []byte(`journal_dir = "`+journalDir+`"`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fake := &run.Fake{}
	return App{Out: io.Discard, Err: io.Discard, Runner: fake, ConfigPath: cfgPath}, journalDir, fake
}

func TestTodayCreatesAndOpens(t *testing.T) {
	app, dir, fake := newApp(t)
	if err := app.Run(context.Background(), []string{"today"}); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	want := filepath.Join(dir, now.Format("2006"), strings.ToLower(now.Format("01-Jan")), strings.ToLower(now.Format("02-Mon"))+".md")
	body, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("entry not created: %v", err)
	}
	if !strings.HasPrefix(string(body), "# ") || !strings.HasSuffix(string(body), "\n\n\n") {
		t.Errorf("unexpected entry body: %q", body)
	}
	if len(fake.Calls) != 1 || fake.Calls[0].Name != "nvim" {
		t.Errorf("editor not launched: %v", fake.Calls)
	}
}

func TestTodayOffsetSelectsPastDay(t *testing.T) {
	app, dir, _ := newApp(t)
	if err := app.Run(context.Background(), []string{"today", "-1"}); err != nil {
		t.Fatal(err)
	}
	y := time.Now().AddDate(0, 0, -1)
	want := filepath.Join(dir, y.Format("2006"), strings.ToLower(y.Format("01-Jan")), strings.ToLower(y.Format("02-Mon"))+".md")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("yesterday's entry not created at %s: %v", want, err)
	}
}

func TestTodayNewCreatesVariants(t *testing.T) {
	app, dir, _ := newApp(t)
	if err := app.Run(context.Background(), []string{"today"}); err != nil {
		t.Fatal(err)
	}
	if err := app.Run(context.Background(), []string{"today", "-n"}); err != nil {
		t.Fatal(err)
	}
	if err := app.Run(context.Background(), []string{"today", "-n"}); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	entryDir := filepath.Join(dir, now.Format("2006"), strings.ToLower(now.Format("01-Jan")))
	base := strings.ToLower(now.Format("02-Mon"))
	for _, name := range []string{base + ".md", base + "2.md", base + "3.md"} {
		if _, err := os.Stat(filepath.Join(entryDir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

func TestCommitStagesAllDayFiles(t *testing.T) {
	app, dir, fake := newApp(t)
	entryDir := filepath.Join(dir, "2026", "06-jun")
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"28-sun.md", "28-sun2.md", "28-sun3.md"} {
		if err := os.WriteFile(filepath.Join(entryDir, name), []byte("# h\n\n\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	target := filepath.Join(entryDir, "28-sun2.md")
	if err := app.Run(context.Background(), []string{"commit", target, "split", "entry"}); err != nil {
		t.Fatal(err)
	}

	var add *run.Call
	for i := range fake.Calls {
		if fake.Calls[i].Name == "git" && len(fake.Calls[i].Args) > 2 && fake.Calls[i].Args[2] == "add" {
			add = &fake.Calls[i]
		}
	}
	if add == nil {
		t.Fatal("no git add call")
	}
	staged := strings.Join(add.Args[3:], " ")
	for _, name := range []string{"28-sun.md", "28-sun2.md", "28-sun3.md"} {
		if !strings.Contains(staged, name) {
			t.Errorf("%s not staged; add args = %v", name, add.Args)
		}
	}

	// The message is appended only to the file being committed.
	body, _ := os.ReadFile(target)
	if !strings.Contains(string(body), "> split entry") {
		t.Errorf("append missing from target: %q", body)
	}
	other, _ := os.ReadFile(filepath.Join(entryDir, "28-sun.md"))
	if strings.Contains(string(other), "> split entry") {
		t.Errorf("message leaked into a sibling file")
	}
}

func TestTodayRejectsBadArg(t *testing.T) {
	app, _, _ := newApp(t)
	if err := app.Run(context.Background(), []string{"today", "garbage"}); err == nil {
		t.Error("expected error for bad argument")
	}
}

func TestCommitAppendsLintsAndCommits(t *testing.T) {
	app, dir, fake := newApp(t)

	// Seed an entry the layout accepts: derive its path from a fixed date.
	d := time.Date(2026, 6, 28, 16, 43, 0, 0, time.UTC)
	entryDir := filepath.Join(dir, "2026", "06-jun")
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(entryDir, "28-sun.md")
	if err := os.WriteFile(file, []byte("# header\n\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = d

	if err := app.Run(context.Background(), []string{"commit", file, "did", "stuff"}); err != nil {
		t.Fatal(err)
	}

	body, _ := os.ReadFile(file)
	if !strings.Contains(string(body), "> did stuff") {
		t.Errorf("message not appended: %q", body)
	}

	var lint, add, commit *run.Call
	for i := range fake.Calls {
		switch {
		case strings.HasSuffix(fake.Calls[i].Name, "markdownlint-cli2"):
			lint = &fake.Calls[i]
		case fake.Calls[i].Name == "git" && fake.Calls[i].Args[2] == "add":
			add = &fake.Calls[i]
		case fake.Calls[i].Name == "git" && fake.Calls[i].Args[2] == "commit":
			commit = &fake.Calls[i]
		}
	}
	if lint == nil || add == nil || commit == nil {
		t.Fatalf("missing steps; calls = %+v", fake.Calls)
	}
	if msg := commit.Args[4]; msg != "2026-06-28 did stuff" {
		t.Errorf("commit message = %q", msg)
	}
}

func TestCommitRequiresFileAndMessage(t *testing.T) {
	app, _, _ := newApp(t)
	if err := app.Run(context.Background(), []string{"commit", "only-file.md"}); err == nil {
		t.Error("expected error when message missing")
	}
}

func TestCommitRejectsNonEntryFile(t *testing.T) {
	app, dir, _ := newApp(t)
	bad := filepath.Join(dir, "notanentry.md")
	if err := os.WriteFile(bad, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := app.Run(context.Background(), []string{"commit", bad, "msg"}); err == nil {
		t.Error("expected error for file that doesn't match the layout")
	}
}

func TestUnknownCommand(t *testing.T) {
	app, _, _ := newApp(t)
	if err := app.Run(context.Background(), []string{"frobnicate"}); err == nil {
		t.Error("expected error for unknown command")
	}
}
