package jsonschema

import (
	"encoding/json"
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
)

// Enum is JSON Schema enum validator description.
type Enum []json.RawMessage

// MarshalYAML implements yaml.Marshaler.
func (n Enum) MarshalYAML() (any, error) {
	var vals []*yaml.Node
	for _, val := range n {
		node, err := convertJSONToRawYAML(val)
		if err != nil {
			return nil, err
		}
		vals = append(vals, node)
	}
	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: vals,
	}, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (n *Enum) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return &yaml.UnmarshalError{
			Node: node,
			Type: reflect.TypeOf(n),
			Err:  errors.Errorf("cannot unmarshal %s into %T", node.ShortTag(), n),
		}
	}
	*n = (*n)[:0]
	for _, val := range node.Content {
		raw, err := convertYAMLtoRawJSON(val)
		if err != nil {
			return err
		}
		*n = append(*n, raw)
	}
	return nil
}
