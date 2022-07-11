package json

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Lines is a sorted slice of newline offsets.
type Lines struct {
	data []byte
	// lines stores newline offsets.
	//
	// idx is the line number (counts from 0).
	lines []int
}

// Search returns the nearest line number to the given offset.
//
// NOTE: may return index bigger than lines length.
func (l Lines) Search(offset int) int {
	// The index is the line number.
	lines := l.lines
	idx := sort.Search(len(lines), func(i int) bool {
		return lines[i] >= offset
	})
	return idx
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

// Location is a JSON value location.
type Location struct {
	Line, Column int
	Node         *yaml.Node
}

func (l *Location) fromNode(node *yaml.Node) {
	*l = Location{
		Line:   node.Line,
		Column: node.Column,
		Node:   node,
	}
}

// Field tries to find the child node using given key and returns its location.
// If such node is not found or parent node is not a mapping, Field returns location of the parent node.
//
// NOTE: child location will point to the value node, not to the key node.
func (l Location) Field(key string) (loc Location) {
	n := l.Node
	if n != nil && n.Kind == yaml.DocumentNode {
		if len(n.Content) < 1 {
			return l
		}
		n = n.Content[0]
	}

	if n == nil || n.Kind != yaml.MappingNode || len(n.Content) < 2 {
		return l
	}

	children := n.Content
	for i := 0; i < len(children); i += 2 {
		keyNode, valueNode := children[i], children[i+1]
		if keyNode.Value == key {
			loc.fromNode(valueNode)
			return loc
		}
	}
	return l
}

// Index tries to find the child node using given index and returns its location.
// If such node is not found or parent node is not a sequence, Field returns location of the parent node.
func (l Location) Index(idx int) (loc Location) {
	n := l.Node
	if n != nil && n.Kind == yaml.DocumentNode {
		if len(n.Content) < 1 {
			return l
		}
		n = n.Content[0]
	}

	if n == nil || n.Kind != yaml.SequenceNode {
		return l
	}

	children := n.Content
	if idx < 0 || idx >= len(children) {
		return l
	}
	loc.fromNode(children[idx])
	return loc
}

// String implements fmt.Stringer.
func (l Location) String() string {
	line, column, n := l.Line, l.Column, l.Node
	if n != nil {
		line, column = n.Line, n.Column
	}
	if column == 0 {
		return strconv.Itoa(line)
	}
	return fmt.Sprintf("%d:%d", line, column)
}

// WithFilename prints the location with the given filename.
//
// If filename is empty, the location is printed as is.
func (l Location) WithFilename(filename string) string {
	if filename != "" {
		filename += ":"
	}
	return filename + l.String()
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

// Field tries to find the child node using given key and returns its location.
//
// See Field method of Location.
func (l Locator) Field(key string) (loc Locator) {
	if l.set {
		loc.SetLocation(l.location.Field(key))
	}
	return
}

// Index tries to find the child node using given index and returns its location.
//
// See Index method of Location.
func (l Locator) Index(idx int) (loc Locator) {
	if l.set {
		loc.SetLocation(l.location.Index(idx))
	}
	return
}

// MarshalYAML implements yaml.Marshaler.
func (l *Locator) MarshalYAML(n *yaml.Node) error {
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (l *Locator) UnmarshalYAML(n *yaml.Node) error {
	var loc Location
	loc.fromNode(n)
	l.SetLocation(loc)
	return nil
}
