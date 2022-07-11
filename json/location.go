package json

import (
	"bytes"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// Lines is a sorted slice of newline offsets.
type Lines struct {
	data []byte
	// lines stores newline offsets.
	//
	// idx is the line number (counts from 0).
	lines []int64
}

// Search returns the nearest line number to the given offset.
//
// NOTE: may return index bigger than lines length.
func (l Lines) Search(offset int64) int64 {
	// The index is the line number.
	lines := l.lines
	idx := sort.Search(len(lines), func(i int) bool {
		return lines[i] >= offset
	})
	return int64(idx)
}

// Line returns offset range of the line.
//
// NOTE: the line number is 1-based. Returns (-1, -1) if the line is invalid.
func (l Lines) Line(n int) (start, end int64) {
	n--
	end = int64(len(l.data))
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
		l.lines = append(l.lines, int64(offset+idx))
		offset += idx + 1
		remain = remain[idx+1:]
	}
}

// LineColumn returns the line and column of the location.
//
// If offset is invalid, line and column are 0 and ok is false.
func (l Lines) LineColumn(offset int64) (line, column int64, ok bool) {
	if offset < 0 || offset >= int64(len(l.data)) {
		return 0, 0, false
	}
	{
		unread := l.data[offset:]
		trimmed := bytes.TrimLeft(unread, "\x20\t\r\n,:")
		if len(trimmed) != len(unread) {
			// Skip leading whitespace, because decoder does not do it.
			offset += int64(len(unread) - len(trimmed))
		}
	}

	line = l.Search(offset)
	if line > 0 {
		var prevLine int64
		if line-1 < int64(len(l.lines)) {
			prevLine = l.lines[line-1]
		}
		column = offset - prevLine
	} else {
		// Offset is on the first line. Column counts from 1.
		column = offset + 1
	}

	// Line counts from 1.
	return line + 1, column, true
}

// Location is a JSON value location.
type Location struct {
	Filename     string
	Line, Column int64
	Node         *yaml.Node
}

// String implements fmt.Stringer.
func (l Location) String() string {
	n := l.Node
	if n == nil {
		return "<empty location>"
	}
	if l.Filename == "" {
		return fmt.Sprintf("%d:%d", n.Line, n.Column)
	}
	return fmt.Sprintf("%s:%d:%d", l.Filename, n.Line, n.Column)
}

// WithFilename creates new Location with the given filename.
func (l Location) WithFilename(filename string) Location {
	return Location{
		Filename: filename,
		Line:     l.Line,
		Column:   l.Column,
		Node:     l.Node,
	}
}

// Locatable is an interface for JSON value location store.
type Locatable interface {
	// SetLocation sets the location of the value.
	SetLocation(Location)

	// Location returns the location of the value if it is set.
	Location() (Location, bool)
}

// Locator stores the Location of a JSON value.
type Locator struct {
	location Location
	set      bool
}

// SetLocation sets the location of the value.
func (l *Locator) SetLocation(loc Location) {
	l.location = loc
	l.set = true
}

// Location returns the location of the value if it is set.
func (l Locator) Location() (Location, bool) {
	return l.location, l.set
}

// MarshalYAML implements yaml.Marshaler.
func (l *Locator) MarshalYAML(n *yaml.Node) error {
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (l *Locator) UnmarshalYAML(n *yaml.Node) error {
	l.SetLocation(Location{
		Line:   int64(n.Line),
		Column: int64(n.Column),
		Node:   n,
	})
	return nil
}
