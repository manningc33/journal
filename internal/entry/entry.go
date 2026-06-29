// Package entry maps between calendar dates and the on-disk layout of journal
// files. It is pure (no I/O) so the date<->path logic is trivially testable.
package entry

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/manningc33/journal/internal/config"
	"github.com/manningc33/journal/internal/datefmt"
)

// Entry is the resolved location of a journal entry for a given date.
type Entry struct {
	Date time.Time
	Dir  string // absolute directory containing the entry
	Path string // absolute path to the entry file
}

// For computes the entry location for date under root.
func For(root string, date time.Time, f config.FormatConfig) Entry {
	rel := f.Render(f.Dir, date)
	name := f.Render(f.File, date)
	dir := filepath.Join(root, filepath.FromSlash(rel))
	return Entry{Date: date, Dir: dir, Path: filepath.Join(dir, name)}
}

// NewContents returns the initial body for a freshly created entry: the
// formatted header followed by two blank lines (cursor opens on line 3).
func (e Entry) NewContents(f config.FormatConfig) string {
	return f.Render(f.Header, e.Date) + "\n\n\n"
}

// DateFromPath derives the entry's date from its path and validates that the
// path actually conforms to the configured layout. Validation is done by
// re-deriving the canonical path from the parsed date and comparing, which
// catches malformed paths regardless of the format in use.
func DateFromPath(root, path string, f config.FormatConfig) (time.Time, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return time.Time{}, err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return time.Time{}, err
	}
	rel, err := filepath.Rel(absRoot, abs)
	if err != nil {
		return time.Time{}, err
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "../") {
		return time.Time{}, fmt.Errorf("%s is outside the journal directory %s", path, root)
	}

	pattern := f.Dir + "/" + f.File

	// Canonical entry: the path parses and round-trips exactly.
	if date, err := datefmt.Parse(pattern, rel); err == nil {
		if filepath.Clean(For(absRoot, date, f).Path) == filepath.Clean(abs) {
			return date, nil
		}
	}

	// Variant entry (created by `today -n`): the filename carries a numeric
	// suffix before the extension. Strip it, parse the canonical form, then
	// confirm the original name really is a variant of the canonical entry.
	relDir, name := splitLast(rel)
	if canon := stripTrailingDigits(name); canon != name {
		if date, err := datefmt.Parse(pattern, relDir+canon); err == nil {
			want := For(absRoot, date, f)
			if _, ok := VariantNum(name, filepath.Base(want.Path)); ok &&
				filepath.Clean(filepath.Dir(want.Path)) == filepath.Clean(filepath.Dir(abs)) {
				return date, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("%s is not a valid journal entry", abs)
}

// VariantName returns the filename of the n-th same-day entry. n<=1 yields the
// canonical name; n>=2 inserts the number before the extension, so
// "28-sun.md" becomes "28-sun2.md". This is the naming `today -n` produces.
func VariantName(canonical string, n int) string {
	if n <= 1 {
		return canonical
	}
	ext := filepath.Ext(canonical)
	return canonical[:len(canonical)-len(ext)] + strconv.Itoa(n) + ext
}

// VariantNum reports whether name is the canonical entry or one of its numeric
// variants, returning the variant number (1 for canonical).
func VariantNum(name, canonical string) (int, bool) {
	if name == canonical {
		return 1, true
	}
	ext := filepath.Ext(canonical)
	stem := canonical[:len(canonical)-len(ext)]
	if !strings.HasPrefix(name, stem) || !strings.HasSuffix(name, ext) {
		return 0, false
	}
	mid := name[len(stem) : len(name)-len(ext)]
	if !allDigits(mid) {
		return 0, false
	}
	if n, err := strconv.Atoi(mid); err == nil && n >= 2 {
		return n, true
	}
	return 0, false
}

// stripTrailingDigits removes a run of digits immediately before the extension.
func stripTrailingDigits(name string) string {
	ext := filepath.Ext(name)
	stem := name[:len(name)-len(ext)]
	i := len(stem)
	for i > 0 && stem[i-1] >= '0' && stem[i-1] <= '9' {
		i--
	}
	return stem[:i] + ext
}

func splitLast(rel string) (dir, name string) {
	if i := strings.LastIndex(rel, "/"); i >= 0 {
		return rel[:i+1], rel[i+1:]
	}
	return "", rel
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
