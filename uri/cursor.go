package uri

import (
	"io"
	"strings"
)

type cursor struct {
	src string
	pos int
}

func (c *cursor) readUntil(until byte) (string, error) {
	from := c.pos
	idx := strings.IndexByte(c.src[from:], until)
	if idx < 0 {
		c.pos = len(c.src)
		return "", io.EOF
	}
	c.pos += idx + 1
	return c.src[from : from+idx], nil
}

func (c *cursor) readValue(sep byte) (v string, hasNext bool, err error) {
	before, _, ok := strings.Cut(c.src[c.pos:], string(sep))
	if !ok {
		if before == "" {
			return "", false, io.EOF
		}
		c.pos = len(c.src)
		return before, false, nil
	}
	c.pos += len(before) + 1
	return before, true, nil
}

func (c *cursor) eat(r byte) bool {
	s := c.src[c.pos:]
	if len(s) > 0 && s[0] == r {
		c.pos++
		return true
	}
	return false
}

func (c *cursor) readAll() (string, error) {
	if c.pos == len(c.src) {
		return "", io.EOF
	}

	defer func() { c.pos = len(c.src) }()
	return c.src[c.pos:], nil
}
