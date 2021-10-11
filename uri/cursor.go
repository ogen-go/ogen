package uri

import "io"

type cursor struct {
	src []rune
	pos int
}

func (c *cursor) readAt(at rune) (string, error) {
	var data []rune
	for {
		r, ok := c.read()
		if !ok {
			return "", io.EOF
		}

		if r == at {
			return string(data), nil
		}

		data = append(data, r)
	}
}

func (c *cursor) readValue(delim rune) (v string, hasNext bool, err error) {
	var data []rune
	for {
		r, ok := c.read()
		if !ok {
			if len(data) == 0 {
				return "", false, io.EOF
			}
			return string(data), false, nil
		}

		if r == delim {
			return string(data), len(c.src) != c.pos, nil
		}

		data = append(data, r)
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
