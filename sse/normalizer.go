package sse

import (
	"bytes"
	"io"
)

// newlineNormalizer converts bare CR line endings to LF while preserving CRLF
// sequences and all other bytes unchanged.
type newlineNormalizer struct {
	r io.Reader

	// prevCR indicates whether the previous read ended with CR.
	prevCR bool
	buf    [1]byte
	// pending stores the output byte that did not fit in the buffer.
	pending    byte
	hasPending bool
}

func (r *newlineNormalizer) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	for {
		if r.hasPending {
			r.hasPending = false
			p[0] = r.pending
			return 1, nil
		}

		if r.prevCR {
			r.prevCR = false
			n, err := r.r.Read(r.buf[:])
			if n == 0 {
				if err == nil {
					return 0, nil
				}
				p[0] = '\n'
				return 1, err
			}

			switch r.buf[0] {
			case '\n':
				p[0] = '\r'
				if len(p) > 1 {
					p[1] = '\n'
					return 2, err
				}
				r.pending = '\n'
			case '\r':
				p[0] = '\n'
				r.prevCR = true
				return 1, nil
			default:
				p[0] = '\n'
				if len(p) > 1 {
					p[1] = r.buf[0]
					return 2, err
				}
				r.pending = r.buf[0]
			}
			r.hasPending = true
			return 1, nil
		}

		n, err := r.r.Read(p)
		if n == 0 {
			return n, err
		}

		i := bytes.IndexByte(p[:n], '\r')
		if i == -1 {
			// Fast path.
			return n, err
		}

		for ; i < n; i++ {
			if p[i] != '\r' {
				continue
			}
			if i+1 < n {
				if p[i+1] != '\n' {
					p[i] = '\n'
				}
				continue
			}
			if err != nil {
				p[i] = '\n'
				return n, err
			}

			r.prevCR = true
			if i > 0 {
				return i, nil
			}
		}
		if r.prevCR {
			continue
		}
		return n, err
	}
}
