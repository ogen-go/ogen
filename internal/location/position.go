package location

import (
	"fmt"
	"strconv"

	"github.com/go-faster/yaml"
)

// Position is a value position.
type Position struct {
	Line, Column int
	Node         *yaml.Node
}

// FromNode sets the position of the value from the given node.
func (p *Position) FromNode(node *yaml.Node) {
	*p = Position{
		Line:   node.Line,
		Column: node.Column,
		Node:   node,
	}
}

func (p Position) mapping() ([]*yaml.Node, bool) {
	n := p.Node
	if n != nil && n.Kind == yaml.DocumentNode {
		if len(n.Content) < 1 {
			return nil, false
		}
		n = n.Content[0]
	}

	if n == nil || n.Kind != yaml.MappingNode || len(n.Content) < 2 {
		return nil, false
	}

	return n.Content, true
}

// Key tries to find the child node using given key and returns its position.
// If such node is not found or parent node is not a mapping, Key returns position of the parent node.
//
// NOTE: child position will point to the key node, not to the value node.
// Use Field if you want position of the value.
func (p Position) Key(key string) (loc Position) {
	children, ok := p.mapping()
	if !ok {
		return loc
	}

	for i := 0; i < len(children); i += 2 {
		keyNode := children[i]
		if keyNode.Value == key {
			loc.FromNode(keyNode)
			return loc
		}
	}
	return p
}

// Field tries to find the child node using given key and returns its position.
// If such node is not found or parent node is not a mapping, Field returns position of the parent node.
//
// NOTE: child position will point to the value node, not to the key node.
// Use Key if you want position of the key.
func (p Position) Field(key string) (loc Position) {
	children, ok := p.mapping()
	if !ok {
		return loc
	}

	for i := 0; i < len(children); i += 2 {
		keyNode, valueNode := children[i], children[i+1]
		if keyNode.Value == key {
			loc.FromNode(valueNode)
			return loc
		}
	}
	return p
}

// Index tries to find the child node using given index and returns its position.
// If such node is not found or parent node is not a sequence, Field returns position of the parent node.
func (p Position) Index(idx int) (loc Position) {
	n := p.Node
	if n != nil && n.Kind == yaml.DocumentNode {
		if len(n.Content) < 1 {
			return p
		}
		n = n.Content[0]
	}

	if n == nil || n.Kind != yaml.SequenceNode {
		return p
	}

	children := n.Content
	if idx < 0 || idx >= len(children) {
		return p
	}
	loc.FromNode(children[idx])
	return loc
}

// String implements fmt.Stringer.
func (p Position) String() string {
	line, column, n := p.Line, p.Column, p.Node
	if n != nil {
		line, column = n.Line, n.Column
	}
	if column == 0 {
		return strconv.Itoa(line)
	}
	return fmt.Sprintf("%d:%d", line, column)
}

// WithFilename prints the position with the given filename.
//
// If filename is empty, the position is printed as is.
func (p Position) WithFilename(filename string) string {
	if filename != "" {
		filename += ":"
	}
	return filename + p.String()
}
