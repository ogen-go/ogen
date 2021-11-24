package uri

import "io"

type cursor struct {
	src []rune
	pos int
}

func (c *cursor) readUntil(until rune) (string, error) {
	var from, to = c.pos, c.pos
	for {
		r, ok := c.read()
		if !ok {
			return "", io.EOF
		}

		if r == until {
			return string(c.src[from:to]), nil
		}
		to++
	}
}

func (c *cursor) readValue(sep rune) (v string, hasNext bool, err error) {
	var from, to = c.pos, c.pos
	for {
		r, ok := c.read()
		if !ok {
			if to-from == 0 {
				return "", false, io.EOF
			}
			return string(c.src[from:to]), false, nil
		}

		if r == sep {
			return string(c.src[from:to]), len(c.src) != c.pos, nil
		}

		to++
	}
}

func (c *cursor) read() (rune, bool) {
	if len(c.src) == c.pos {
		return rune(0), false
	}

	defer func() { c.pos++ }()
	return c.src[c.pos], true
}

func (c *cursor) eat(r rune) bool {
	rr, ok := c.read()
	if !ok {
		return false
	}

	if r != rr {
		c.pos--
		return false
	}

	return true
}

func (c *cursor) readAll() (string, error) {
	if c.pos == len(c.src) {
		return "", io.EOF
	}

	defer func() { c.pos = len(c.src) }()
	return string(c.src[c.pos:]), nil
}
