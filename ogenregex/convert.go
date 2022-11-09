package ogenregex

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-faster/errors"
	"golang.org/x/text/unicode/rangetable"
)

// Copied from dop251/goja, to avoid dependency.
//
// All rights belong to the original author.
//
// https://github.com/dop251/goja/blob/3b8a68ca89b4fa7086a4236695032e10a69b2472/parser/regexp.go#L58

const (
	whitespaceChars = " \f\n\r\t\v" +
		"\u00a0\u1680" +
		"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a" +
		"\u2028\u2029" +
		"\u202f\u205f" +
		"\u3000\ufeff"
	re2Dot = "[^\r\n\u2028\u2029]"
)

func digitValue(chr rune) int {
	switch {
	case '0' <= chr && chr <= '9':
		return int(chr - '0')
	case 'a' <= chr && chr <= 'f':
		return int(chr - 'a' + 10)
	case 'A' <= chr && chr <= 'F':
		return int(chr - 'A' + 10)
	}
	return 16 // Larger than any legal digit value
}

var (
	unicodeRangeIDNeg      = rangetable.Merge(unicode.Pattern_Syntax, unicode.Pattern_White_Space)
	unicodeRangeIDStartPos = rangetable.Merge(unicode.Letter, unicode.Nl, unicode.Other_ID_Start)
	unicodeRangeIDContPos  = rangetable.Merge(
		unicodeRangeIDStartPos,
		unicode.Mn,
		unicode.Mc,
		unicode.Nd,
		unicode.Pc,
		unicode.Other_ID_Continue,
	)
)

func isIDPartUnicode(r rune) bool {
	return unicode.Is(unicodeRangeIDContPos, r) && !unicode.Is(unicodeRangeIDNeg, r) || r == '\u200C' || r == '\u200D'
}

func isIdentifierPart(chr rune) bool {
	return chr == '$' || chr == '_' || chr == '\\' ||
		'a' <= chr && chr <= 'z' || 'A' <= chr && chr <= 'Z' ||
		'0' <= chr && chr <= '9' ||
		chr >= utf8.RuneSelf && isIDPartUnicode(chr)
}

// Convert converts a ECMA-262 regular expression to Go's regular expression.
//
// If the conversion is not possible, ("", false) is returned.
func Convert(pattern string) (string, bool) {
	if pattern == "" {
		return "", true
	}

	p := parser{
		str:    pattern,
		length: len(pattern),
	}
	if err := p.parse(); err != nil {
		return "", false
	}

	return p.ResultString(), true
}

type parser struct {
	str    string
	length int

	chr       rune // The current character
	chrOffset int  // The offset of current character
	offset    int  // The offset after current character (may be greater than 1)

	err error

	goRegexp   strings.Builder
	passOffset int
}

func (p *parser) ResultString() string {
	if p.passOffset != -1 {
		return p.str[:p.passOffset]
	}
	return p.goRegexp.String()
}

func (p *parser) parse() error {
	p.read() // Pull in the first character
	p.scan()
	return p.err
}

func (p *parser) read() {
	if p.offset < p.length {
		p.chrOffset = p.offset
		chr, width := rune(p.str[p.offset]), 1
		if chr >= utf8.RuneSelf { // !ASCII
			chr, width = utf8.DecodeRuneInString(p.str[p.offset:])
			if chr == utf8.RuneError && width == 1 {
				p.error(true, "Invalid UTF-8 character")
				return
			}
		}
		p.offset += width
		p.chr = chr
	} else {
		p.chrOffset = p.length
		p.chr = -1 // EOF
	}
}

func (p *parser) stopPassing() {
	p.goRegexp.Grow(3 * len(p.str) / 2)
	p.goRegexp.WriteString(p.str[:p.passOffset])
	p.passOffset = -1
}

func (p *parser) write(data []byte) {
	if p.passOffset != -1 {
		p.stopPassing()
	}
	p.goRegexp.Write(data)
}

func (p *parser) writeByte(b byte) {
	if p.passOffset != -1 {
		p.stopPassing()
	}
	p.goRegexp.WriteByte(b)
}

func (p *parser) writeString(s string) {
	if p.passOffset != -1 {
		p.stopPassing()
	}
	p.goRegexp.WriteString(s)
}

func (p *parser) scan() {
	for p.chr != -1 {
		switch p.chr {
		case '\\':
			p.read()
			p.scanEscape(false)
		case '(':
			p.pass()
			p.scanGroup()
		case '[':
			p.scanBracket()
		case ')':
			p.error(true, "Unmatched ')'")
			return
		case '.':
			p.writeString(re2Dot)
			p.read()
		default:
			p.pass()
		}
	}
}

// (...)
func (p *parser) scanGroup() {
	str := p.str[p.chrOffset:]
	if len(str) > 1 { // A possibility of (?= or (?!
		if str[0] == '?' {
			ch := str[1]
			switch {
			case ch == '=' || ch == '!':
				p.error(false, "re2: Invalid (%s) <lookahead>", p.str[p.chrOffset:p.chrOffset+2])
				return
			case ch == '<':
				p.error(false, "re2: Invalid (%s) <lookbehind>", p.str[p.chrOffset:p.chrOffset+2])
				return
			case ch != ':':
				p.error(true, "Invalid group")
				return
			}
		}
	}
	for p.chr != -1 && p.chr != ')' {
		switch p.chr {
		case '\\':
			p.read()
			p.scanEscape(false)
		case '(':
			p.pass()
			p.scanGroup()
		case '[':
			p.scanBracket()
		case '.':
			p.writeString(re2Dot)
			p.read()
		default:
			p.pass()
			continue
		}
	}
	if p.chr != ')' {
		p.error(true, "Unterminated group")
		return
	}
	p.pass()
}

// [...]
func (p *parser) scanBracket() {
	str := p.str[p.chrOffset:]
	if strings.HasPrefix(str, "[]") {
		// [] -- Empty character class
		p.writeString("[^\u0000-\U0001FFFF]")
		p.offset++
		p.read()
		return
	}

	if strings.HasPrefix(str, "[^]") {
		p.writeString("[\u0000-\U0001FFFF]")
		p.offset += 2
		p.read()
		return
	}

	p.pass()
	for p.chr != -1 {
		if p.chr == ']' {
			break
		} else if p.chr == '\\' {
			p.read()
			p.scanEscape(true)
			continue
		}
		p.pass()
	}
	if p.chr != ']' {
		p.error(true, "Unterminated character class")
		return
	}
	p.pass()
}

// \...
func (p *parser) scanEscape(inClass bool) {
	offset := p.chrOffset

	var length, base uint32
	switch p.chr {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		var value int64
		size := 0
		for {
			digit := int64(digitValue(p.chr))
			if digit >= 8 {
				// Not a valid digit
				break
			}
			value = value*8 + digit
			p.read()
			size++
		}
		if size == 1 { // The number of characters read
			if value != 0 {
				// An invalid backreference
				p.error(false, "re2: Invalid \\%d <backreference>", value)
				return
			}
			p.passString(offset-1, p.chrOffset)
			return
		}
		tmp := []byte{'\\', 'x', '0', 0}
		if value >= 16 {
			tmp = tmp[0:2]
		} else {
			tmp = tmp[0:3]
		}
		tmp = strconv.AppendInt(tmp, value, 16)
		p.write(tmp)
		return

	case '8', '9':
		p.read()
		p.error(false, "re2: Invalid \\%s <backreference>", p.str[offset:p.chrOffset])
		return

	case 'x':
		p.read()
		length, base = 2, 16

	case 'u':
		p.read()
		if p.chr == '{' {
			p.read()
			length, base = 0, 16
		} else {
			length, base = 4, 16
		}

	case 'b':
		if inClass {
			p.write([]byte{'\\', 'x', '0', '8'})
			p.read()
			return
		}
		fallthrough

	case 'B':
		fallthrough

	case 'd', 'D', 'w', 'W':
		// This is slightly broken, because ECMAScript
		// includes \v in \s, \S, while re2 does not
		fallthrough

	case '\\':
		fallthrough

	case 'f', 'n', 'r', 't', 'v':
		p.passString(offset-1, p.offset)
		p.read()
		return

	case 'c':
		p.read()
		var value int64
		switch {
		case 'a' <= p.chr && p.chr <= 'z':
			value = int64(p.chr - 'a' + 1)
		case 'A' <= p.chr && p.chr <= 'Z':
			value = int64(p.chr - 'A' + 1)
		default:
			p.writeByte('c')
			return
		}
		tmp := []byte{'\\', 'x', '0', 0}
		if value >= 16 {
			tmp = tmp[0:2]
		} else {
			tmp = tmp[0:3]
		}
		tmp = strconv.AppendInt(tmp, value, 16)
		p.write(tmp)
		p.read()
		return
	case 's':
		if inClass {
			p.writeString(whitespaceChars)
		} else {
			p.writeString("[" + whitespaceChars + "]")
		}
		p.read()
		return
	case 'S':
		if inClass {
			p.error(false, "S in class")
			return
		}
		p.writeString("[^" + whitespaceChars + "]")
		p.read()
		return
	default:
		// $ is an identifier character, so we have to have
		// a special case for it here
		if p.chr == '$' || p.chr < utf8.RuneSelf && !isIdentifierPart(p.chr) {
			// A non-identifier character needs escaping
			p.passString(offset-1, p.offset)
			p.read()
			return
		}
		// Unescape the character for re2
		p.pass()
		return
	}

	// Otherwise, we're a \u.... or \x...
	valueOffset := p.chrOffset

	if length > 0 {
		for length := length; length > 0; length-- {
			digit := uint32(digitValue(p.chr))
			if digit >= base {
				// Not a valid digit
				goto skip
			}
			p.read()
		}
	} else {
		for p.chr != '}' && p.chr != -1 {
			digit := uint32(digitValue(p.chr))
			if digit >= base {
				// Not a valid digit
				goto skip
			}
			p.read()
		}
	}

	switch length {
	case 0, 4:
		p.write([]byte{'\\', 'x', '{'})
		p.passString(valueOffset, p.chrOffset)
		if length != 0 {
			p.writeByte('}')
		}
	case 2:
		p.passString(offset-1, valueOffset+2)
	default:
		// Should never, ever get here...
		p.error(true, "re2: Illegal branch in scanEscape")
		return
	}

	return
skip:
	p.passString(offset, p.chrOffset)
}

func (p *parser) pass() {
	if p.passOffset == p.chrOffset {
		p.passOffset = p.offset
	} else {
		if p.passOffset != -1 {
			p.stopPassing()
		}
		if p.chr != -1 {
			p.goRegexp.WriteRune(p.chr)
		}
	}
	p.read()
}

func (p *parser) passString(start, end int) {
	if p.passOffset == start {
		p.passOffset = end
		return
	}
	if p.passOffset != -1 {
		p.stopPassing()
	}
	p.goRegexp.WriteString(p.str[start:end])
}

func (p *parser) error(fatal bool, format string, args ...interface{}) {
	if p.err != nil {
		return
	}
	p.err = errors.Errorf(format, args...)
	if fatal {
		p.err = errors.Wrap(p.err, "syntax")
	}
	p.offset = p.length
	p.chr = -1
}
