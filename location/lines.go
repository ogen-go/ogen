package location

import (
	"bytes"
)

// Lines is a sorted slice of newline offsets.
type Lines struct {
	data []byte
	// lines stores newline offsets.
	//
	// idx is the line number (counts from 0).
	lines []int
}

// IsZero returns true if lines has zero value.
func (l Lines) IsZero() bool {
	s := struct {
		data  []byte
		lines []int
	}(l)
	return s.data == nil && s.lines == nil
}

// Line returns offset range of the line.
//
// NOTE: the line number is 1-based. Returns (-1, -1) if the line is invalid.
func (l Lines) Line(n int) (start, end int) {
	n--
	end = len(l.data)
	switch {
	case n < 0:
		// Line 0 is invalid.
		return -1, -1
	case n >= len(l.lines):
		// Last line.
		if len(l.lines) > 0 {
			start = l.lines[len(l.lines)-1]
		}
		return start, end
	default:
		if n > 0 {
			start = l.lines[n-1]
		}
		end = l.lines[n]
		return start, end
	}
}

// Collect fills the given slice with the offset of newlines.
func (l *Lines) Collect(data []byte) {
	l.data = data
	l.lines = l.lines[:0]

	var (
		// Remaining data to process.
		remain = data
		// Absolute offset of the current line.
		offset = 0
	)
	for {
		idx := bytes.IndexByte(remain, '\n')
		if idx < 0 {
			break
		}
		l.lines = append(l.lines, offset+idx)
		offset += idx + 1
		remain = remain[idx+1:]
	}
}
