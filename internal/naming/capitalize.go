package naming

import (
	"unicode"
	"unicode/utf8"
)

// Capitalize converts first character to upper.
//
// If the string is invalid UTF-8 or empty, it is returned as is.
func Capitalize(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}
