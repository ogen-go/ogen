package jsonschema

import (
	"github.com/go-json-experiment/json"
	"gopkg.in/yaml.v3"
)

type (
	// RawValue is a raw JSON value.
	RawValue json.RawValue
	// Default is a default value.
	Default = RawValue
	// Example is an example value.
	Example = RawValue
)

// UnmarshalYAML implements yaml.Unmarshaler.
func (n *RawValue) UnmarshalYAML(node *yaml.Node) error {
	raw, err := convertYAMLtoRawJSON(node)
	if err != nil {
		return err
	}
	*n = RawValue(raw)
	return nil
}

// MarshalNextJSON implements json.MarshalerV2.
func (n RawValue) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	val := json.RawValue(n)
	return opts.MarshalNext(e, val)
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (n *RawValue) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	val, err := d.ReadValue()
	if err != nil {
		return err
	}
	*n = append((*n)[:0], val...)
	return nil
}

func convertYAMLtoRawJSON(node *yaml.Node) (json.RawValue, error) {
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
