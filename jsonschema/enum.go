package jsonschema

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"gopkg.in/yaml.v3"
)

// Enum is JSON Schema enum validator description.
type Enum []json.RawMessage

// MarshalYAML implements yaml.Marshaler.
func (n Enum) MarshalYAML() (interface{}, error) {
	var vals []*yaml.Node
	for _, val := range n {
		node, err := convertJSONToRawYAML(val)
		if err != nil {
			return nil, err
		}
		vals = append(vals, &node)
	}
	return vals, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (n *Enum) UnmarshalYAML(node *yaml.Node) error {
	if node.Tag != "!!seq" {
		return errors.Errorf("unexpected tag %s", node.Tag)
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
