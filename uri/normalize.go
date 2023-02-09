package uri

import (
	"strings"
)

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func asciiToUpper(c byte) byte {
	if c >= 'a' && c <= 'f' {
		return c - ('a' - 'A')
	}
	return c
}

func asciiIsLowercase(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Return true if the specified character should be escaped when
// appearing in a URL path string, according to RFC 3986.
func shouldEscapePath(c byte) bool {
	// §2.3 Unreserved characters (alpha)
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' {
		return false
	}
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // 0-9
		return false
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false
	default:
		// Everything else must be escaped.
		return true
	}
}

// NormalizeEscapedPath normalizes escaped path.
//
// All percent-encoded characters are upper-cased. If s contains unnecessarily escaped
// characters, they are unescaped.
//
// If s contains invalid escape sequence, it returns empty string and false.
func NormalizeEscapedPath(s string) (string, bool) {
	// Search % with lower case octets.
	iter := s
	for {
		idx := strings.IndexByte(iter, '%')
		if idx < 0 {
			return s, true
		}
		if idx+2 >= len(iter) || !ishex(iter[idx+1]) || !ishex(iter[idx+2]) {
			// Invalid escape sequence.
			return "", false
		}
		a, b := iter[idx+1], iter[idx+2]
		if asciiIsLowercase(a) || asciiIsLowercase(b) {
			goto slow
		}
		// Unescape character.
		ch := unhex(a)<<4 | unhex(b)
		if !shouldEscapePath(ch) {
			// Unescape character.
			goto slow
		}
		iter = iter[idx+3:]
	}

slow:
	var t strings.Builder
	t.Grow(len(s))
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			// Unescape character.
			a, b := s[i+1], s[i+2]
			ch := unhex(a)<<4 | unhex(b)
			if shouldEscapePath(ch) {
				t.WriteByte('%')
				t.WriteByte(asciiToUpper(a))
				t.WriteByte(asciiToUpper(b))
			} else {
				t.WriteByte(ch)
			}
			i += 3
		default:
			t.WriteByte(s[i])
			i++
		}
	}
	return t.String(), true
}
