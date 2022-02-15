// Package capitalize contains capitalize function.
package capitalize

import (
	"unicode"
	"unicode/utf8"
)

// Capitalize converts first character to upper.
func Capitalize(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
