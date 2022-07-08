package jsonschema

import (
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"
	"gopkg.in/yaml.v3"
)

// Num represents JSON number.
type Num json.RawValue

// MarshalNextJSON implements json.MarshalerV2.
func (n Num) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	val := json.RawValue(n)
	return opts.MarshalNext(e, val)
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (n *Num) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	offset := d.InputOffset()
	// Check Kind for invalid, next call will return error.
	if kind := d.PeekKind(); kind != '0' && kind != 0 {
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: d.StackPointer(),
			JSONKind:    kind,
			GoType:      reflect.TypeOf(n),
			Err:         errors.Errorf("unexpected type %s", kind.String()),
		}
	}

	val, err := d.ReadValue()
	if err != nil {
		return err
	}
	*n = append((*n)[:0], val...)
	return nil
}

// MarshalJSON implements json.MarshalerV1.
func (n *Num) MarshalJSON() ([]byte, error) {
	// Backward-compatibility with v1.
	return json.Marshal(n)
}

// UnmarshalJSON implements json.UnmarshalerV1.
func (n *Num) UnmarshalJSON(data []byte) error {
	// Backward-compatibility with v1.
	return json.Unmarshal(data, n)
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (n *Num) UnmarshalYAML(node *yaml.Node) error {
	if t := node.Tag; t != "!!int" && t != "!!float" {
		return errors.Errorf("unexpected tag %s", t)
	}
	val, err := convertYAMLtoRawJSON(node)
	if err != nil {
		return err
	}
	return json.Unmarshal(val, n)
}
