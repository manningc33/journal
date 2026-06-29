package entry

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/manningc33/journal/internal/config"
)

func date() time.Time { return time.Date(2026, 6, 28, 16, 43, 0, 0, time.UTC) }

func TestForLayout(t *testing.T) {
	f := config.Defaults().Format
	e := For("/root", date(), f)
	if want := filepath.FromSlash("/root/2026/06-jun"); e.Dir != want {
		t.Errorf("dir = %q, want %q", e.Dir, want)
	}
	if want := filepath.FromSlash("/root/2026/06-jun/28-sun.md"); e.Path != want {
		t.Errorf("path = %q, want %q", e.Path, want)
	}
}

func TestNewContents(t *testing.T) {
	f := config.Defaults().Format
	got := For("/root", date(), f).NewContents(f)
	if want := "# june 28, 2026 (sunday 16:43)\n\n\n"; got != want {
		t.Errorf("contents = %q, want %q", got, want)
	}
}

func TestDateFromPathRoundTrips(t *testing.T) {
	f := config.Defaults().Format
	root := t.TempDir()
	path := filepath.Join(root, "2026", "06-jun", "28-sun.md")
	got, err := DateFromPath(root, path, f)
	if err != nil {
		t.Fatal(err)
	}
	if got.Format("2006-01-02") != "2026-06-28" {
		t.Errorf("date = %s", got.Format("2006-01-02"))
	}
}

func TestVariantName(t *testing.T) {
	cases := map[int]string{1: "28-sun.md", 2: "28-sun2.md", 3: "28-sun3.md", 10: "28-sun10.md"}
	for n, want := range cases {
		if got := VariantName("28-sun.md", n); got != want {
			t.Errorf("VariantName(n=%d) = %q, want %q", n, got, want)
		}
	}
}

func TestVariantNum(t *testing.T) {
	cases := []struct {
		name string
		n    int
		ok   bool
	}{
		{"28-sun.md", 1, true},
		{"28-sun2.md", 2, true},
		{"28-sun10.md", 10, true},
		{"28-mon.md", 0, false},  // different day
		{"28-sun.txt", 0, false}, // different extension
		{"28-sunx.md", 0, false}, // non-numeric suffix
	}
	for _, c := range cases {
		if n, ok := VariantNum(c.name, "28-sun.md"); n != c.n || ok != c.ok {
			t.Errorf("VariantNum(%q) = (%d,%v), want (%d,%v)", c.name, n, ok, c.n, c.ok)
		}
	}
}

func TestDateFromPathVariant(t *testing.T) {
	f := config.Defaults().Format
	root := t.TempDir()
	path := filepath.Join(root, "2026", "06-jun", "28-sun2.md")
	got, err := DateFromPath(root, path, f)
	if err != nil {
		t.Fatal(err)
	}
	if got.Format("2006-01-02") != "2026-06-28" {
		t.Errorf("date = %s", got.Format("2006-01-02"))
	}
}

func TestDateFromPathRejectsWrongWeekday(t *testing.T) {
	f := config.Defaults().Format
	root := t.TempDir()
	// 28th June 2026 is a Sunday; "mon" is inconsistent.
	path := filepath.Join(root, "2026", "06-jun", "28-mon.md")
	if _, err := DateFromPath(root, path, f); err == nil {
		t.Error("expected error for inconsistent weekday")
	}
}

func TestDateFromPathRejectsOutsideRoot(t *testing.T) {
	f := config.Defaults().Format
	if _, err := DateFromPath(t.TempDir(), "/elsewhere/2026/06-jun/28-sun.md", f); err == nil {
		t.Error("expected error for path outside root")
	}
}
