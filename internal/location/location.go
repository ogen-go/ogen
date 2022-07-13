package location

import (
	"fmt"
	"strconv"

	yaml "github.com/go-faster/yamlx"
)

// Location is a JSON value location.
type Location struct {
	Line, Column int
	Node         *yaml.Node
}

// FromNode sets the location of the value from the given node.
func (l *Location) FromNode(node *yaml.Node) {
	*l = Location{
		Line:   node.Line,
		Column: node.Column,
		Node:   node,
	}
}

// Key tries to find the child node using given key and returns its location.
// If such node is not found or parent node is not a mapping, Key returns location of the parent node.
//
// NOTE: child location will point to the key node, not to the value node.
// Use Field if you want location of the value.
func (l Location) Key(key string) (loc Location) {
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
		keyNode := children[i]
		if keyNode.Value == key {
			loc.FromNode(keyNode)
			return loc
		}
	}
	return l
}

// Field tries to find the child node using given key and returns its location.
// If such node is not found or parent node is not a mapping, Field returns location of the parent node.
//
// NOTE: child location will point to the value node, not to the key node.
// Use Key if you want location of the key.
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
			loc.FromNode(valueNode)
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
	loc.FromNode(children[idx])
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

// Key tries to find the child node using given key and returns its location.
//
// See Key method of Location.
func (l Locator) Key(key string) (loc Locator) {
	if l.set {
		loc.SetLocation(l.location.Key(key))
	}
	return
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
	loc.FromNode(n)
	l.SetLocation(loc)
	return nil
}
