package jsonschema

import (
	"encoding/json"

	yaml "github.com/go-faster/yamlx"
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
func (n RawValue) MarshalYAML() (interface{}, error) {
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

func convertJSONToRawYAML(raw json.RawMessage) (node yaml.Node, err error) {
	err = yaml.Unmarshal(raw, &node)
	return node, err
}

func convertYAMLtoRawJSON(node *yaml.Node) (json.RawMessage, error) {
	var tmp interface{}
	if err := node.Decode(&tmp); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(tmp)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
