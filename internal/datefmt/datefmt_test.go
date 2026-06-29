package datefmt

import (
	"testing"
	"time"
)

func TestToLayout(t *testing.T) {
	cases := map[string]string{
		"%Y/%m-%b":               "2006/01-Jan",
		"%d-%a.md":               "02-Mon.md",
		"# %B %e, %Y (%A %H:%M)": "# January _2, 2006 (Monday 15:04)",
		"%Y-%m-%d":               "2006-01-02",
		"literal %z passthru":    "literal %z passthru",
		"100%%":                  "100%",
	}
	for in, want := range cases {
		if got := ToLayout(in); got != want {
			t.Errorf("ToLayout(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatMatchesOriginalScripts(t *testing.T) {
	d := time.Date(2026, 6, 28, 16, 43, 0, 0, time.UTC)
	if got := Format("%Y/%m-%b", d); got != "2026/06-Jun" {
		t.Errorf("dir = %q", got)
	}
	if got := Format("%d-%a.md", d); got != "28-Sun.md" {
		t.Errorf("file = %q", got)
	}
	if got := Format("%Y-%m-%d", d); got != "2026-06-28" {
		t.Errorf("commit = %q", got)
	}
}

func TestParseRoundTrip(t *testing.T) {
	d, err := Parse("%Y/%m-%b/%d-%a.md", "2026/06-jun/28-sun.md") // lowercased on-disk form
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Format("2006-01-02"); got != "2026-06-28" {
		t.Errorf("parsed %q, want 2026-06-28", got)
	}
}
