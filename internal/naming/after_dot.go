package naming

import "strings"

// AfterDot returns the part of the string after the first dot.
//
// If there is no dot in the string or dot is the first character, the whole string is returned.
func AfterDot(s string) string {
	if before, after, ok := strings.Cut(s, "."); ok && before != "" {
		return after
	}
	return s
}
