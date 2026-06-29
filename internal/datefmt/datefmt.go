// Package datefmt translates strftime-style patterns (the same ones the
// original shell scripts fed to date(1)) into Go reference layouts. Reusing a
// single converter for both formatting and parsing keeps the round trip between
// a date and its on-disk path exact.
package datefmt

import (
	"strings"
	"time"
)

// tokens maps a strftime conversion specifier to its Go layout equivalent.
// Only the specifiers used by the journal formats are supported; unknown
// specifiers are passed through verbatim so misconfiguration is visible rather
// than silently dropped.
var tokens = map[byte]string{
	'Y': "2006",    // 4-digit year
	'y': "06",      // 2-digit year
	'm': "01",      // zero-padded month
	'b': "Jan",     // abbreviated month
	'B': "January", // full month
	'd': "02",      // zero-padded day
	'e': "_2",      // space-padded day
	'a': "Mon",     // abbreviated weekday
	'A': "Monday",  // full weekday
	'H': "15",      // 24-hour
	'I': "03",      // 12-hour
	'M': "04",      // minute
	'S': "05",      // second
	'p': "PM",      // AM/PM
	'j': "002",     // day of year
	'%': "%",       // literal percent
}

// ToLayout converts a strftime pattern into a Go reference layout.
func ToLayout(pattern string) string {
	var b strings.Builder
	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '%' || i+1 >= len(pattern) {
			b.WriteByte(pattern[i])
			continue
		}
		i++
		if repl, ok := tokens[pattern[i]]; ok {
			b.WriteString(repl)
		} else {
			b.WriteByte('%')
			b.WriteByte(pattern[i])
		}
	}
	return b.String()
}

// Format renders t using a strftime pattern.
func Format(pattern string, t time.Time) string {
	return t.Format(ToLayout(pattern))
}

// Parse extracts a time from value using a strftime pattern. Month and weekday
// names are matched case-insensitively, so lowercased on-disk paths parse fine.
func Parse(pattern, value string) (time.Time, error) {
	return time.Parse(ToLayout(pattern), value)
}
