package http

import "path"

// MatchContentType returns true if value matches path pattern.
func MatchContentType(pattern, value string) bool {
	ok, _ := path.Match(pattern, value)
	return ok
}
