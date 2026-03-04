package ir

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ogen-go/ogen/internal/naming"
)

// Global line limit, can be configured via SetLineLimit.
var lineLimit = 100

// Global flag to enable or disable pretty documentation.
var prettyDocEnabled = true

// SetLineLimit sets the maximum width of a comment line before it is wrapped.
// Use a negative value to disable line wrapping altogether.
func SetLineLimit(limit int) {
	if limit == 0 {
		// Use default value for zero.
		lineLimit = 100
		return
	}
	lineLimit = limit
}

// GetLineLimit returns the current line limit value.
func GetLineLimit() int {
	return lineLimit
}

// SetPrettyDoc enables or disables pretty documentation.
func SetPrettyDoc(enabled bool) {
	prettyDocEnabled = enabled
}

// IsPrettyDocEnabled returns whether pretty documentation is enabled.
func IsPrettyDocEnabled() bool {
	return prettyDocEnabled
}

func splitLine(s string, limit int) (r []string) {
	// If limit is negative, don't split lines.
	if limit < 0 {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		return []string{s}
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// TODO(tdakkota): handle links in docs
	for {
		if len(s) < limit {
			r = append(r, s)
			return
		}

		idx := strings.LastIndexFunc(s[:limit-1], func(r rune) bool {
			return unicode.IsSpace(r) || r == '.' || r == ',' || r == ';'
		})
		if idx < 0 || len(s)-1 == idx {
			r = append(r, s)
			return
		}

		if ch, size := utf8.DecodeRuneInString(s[idx:]); unicode.IsSpace(ch) {
			r = append(r, s[:idx])
			s = s[idx+size:]
		} else {
			// Do not cut dots and commas.
			r = append(r, s[:idx+size])
			s = s[idx+size:]
		}
	}
}

func prettyDoc(s, deprecation string) (r []string) {
	// If pretty documentation is disabled, return the comment as-is
	if !prettyDocEnabled {
		// Just split by newlines.
		for _, line := range strings.Split(s, "\n") {
			r = append(r, line)
		}

		// Add deprecation notice if provided
		if deprecation != "" {
			if len(r) > 0 {
				// Insert empty line between description and deprecated notice.
				r = append(r, "")
			}
			r = append(r, deprecation)
		}

		return r
	}

	// Original pretty documentation behavior
	// TODO(tdakkota): basic common mark rendering?
	for _, line := range strings.Split(s, "\n") {
		r = append(r, splitLine(line, lineLimit)...)
	}
	if len(r) > 0 {
		r[0] = naming.Capitalize(r[0])

		if last := r[len(r)-1]; len(last) > 0 && last[len(last)-1] != '.' {
			r[len(r)-1] = last + "."
		}
	}
	if deprecation != "" {
		if len(r) > 0 {
			// Insert empty line between description and deprecated notice.
			r = append(r, "")
		}
		r = append(r, deprecation)
	}

	return r
}
