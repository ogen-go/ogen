package uri

import (
	"strings"
)

const cookieEscaper = '%'

var cookieEscapeChars = [128]byte{
	'\x00': 1,
	'\x01': 1,
	'\x02': 1,
	'\x03': 1,
	'\x04': 1,
	'\x05': 1,
	'\x06': 1,
	'\a':   1,
	'\b':   1,
	'\t':   1,
	'\n':   1,
	'\v':   1,
	'\f':   1,
	'\r':   1,
	'\x0e': 1,
	'\x0f': 1,
	'\x10': 1,
	'\x11': 1,
	'\x12': 1,
	'\x13': 1,
	'\x14': 1,
	'\x15': 1,
	'\x16': 1,
	'\x17': 1,
	'\x18': 1,
	'\x19': 1,
	'\x1a': 1,
	'\x1b': 1,
	'\x1c': 1,
	'\x1d': 1,
	'\x1e': 1,
	'\x1f': 1,
	' ':    1,
	'"':    1,
	',':    1,
	';':    1,
	'\\':   1,
	'\x7f': 1,

	// Escape the escape character itself.
	cookieEscaper: 1,
}

const hex = "0123456789ABCDEF"

func escapeCookie(s string) string {
	const length = byte(len(cookieEscapeChars))

	n := 0
	for _, c := range []byte(s) {
		if c >= length || cookieEscapeChars[c] == 1 {
			n++
		}
	}

	// No need to escape.
	if n == 0 {
		return s
	}

	var sb strings.Builder
	// Every escaped char is 2 bytes longer: percent sign and 2 hex digits minus existing byte.
	sb.Grow(len(s) + 2*n)

	for _, c := range []byte(s) {
		if c >= length || cookieEscapeChars[c] == 1 {
			sb.WriteByte(cookieEscaper)
			sb.WriteByte(hex[c>>4])
			sb.WriteByte(hex[c&15])
		} else {
			sb.WriteByte(c)
		}
	}

	return sb.String()
}

func unescapeCookie(s string) (string, bool) {
	n := 0
	for i := 0; i < len(s); {
		if c := s[i]; c == cookieEscaper {
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				return "", false
			}
			n++
			i += 3
		} else {
			i++
		}
	}

	// No need to unescape.
	if n == 0 {
		return s, true
	}

	var sb strings.Builder
	// Every escaped char is 2 bytes longer: percent sign and 2 hex digits minus existing byte.
	sb.Grow(len(s) - 2*n)

	for i := 0; i < len(s); {
		if c := s[i]; c == cookieEscaper {
			sb.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
			i += 3
		} else {
			sb.WriteByte(c)
			i++
		}
	}

	return sb.String(), true
}
