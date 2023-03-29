package jsonschema

import (
	"encoding/json"

	helperyaml "github.com/ghodss/yaml"
	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
)

type (
	// RawValue is a raw JSON value.
	RawValue json.RawMessage
	// Default is a default value.
	Default = RawValue
	// Example is an example value.
	Example = RawValue
)

// MarshalYAML implements yaml.Marshaler.
func (n RawValue) MarshalYAML() (any, error) {
	return convertJSONToRawYAML(json.RawMessage(n))
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (n *RawValue) UnmarshalYAML(node *yaml.Node) error {
	raw, err := convertYAMLtoRawJSON(node)
	if err != nil {
		return err
	}
	*n = RawValue(raw)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (n RawValue) MarshalJSON() ([]byte, error) {
	return json.RawMessage(n).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *RawValue) UnmarshalJSON(b []byte) error {
	*n = append((*n)[:0], b...)
	return nil
}

func convertJSONToRawYAML(raw json.RawMessage) (_ *yaml.Node, rerr error) {
	defer func() {
		if rerr != nil {
			rerr = errors.Wrap(rerr, "convert JSON to YAML")
		}
	}()
	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	if node.Kind == yaml.DocumentNode {
		return node.Content[0], nil
	}
	return &node, nil
}

func convertYAMLtoRawJSON(node *yaml.Node) (_ json.RawMessage, rerr error) {
	defer func() {
		if rerr != nil {
			rerr = errors.Wrap(rerr, "convert YAML to JSON")
		}
	}()
	raw, err := yaml.Marshal(node)
	if err != nil {
		return nil, err
	}
	return helperyaml.YAMLToJSON(raw)
}
