package jsonschema

import (
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"
	"gopkg.in/yaml.v3"

	ogenjson "github.com/ogen-go/ogen/json"
)

// Enum is JSON Schema enum validator description.
type Enum []json.RawValue

// MarshalNextJSON implements json.MarshalerV2.
func (n Enum) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	if err := e.WriteToken(json.ArrayStart); err != nil {
		return err
	}
	for _, val := range n {
		if err := opts.MarshalNext(e, val); err != nil {
			return err
		}
	}
	if err := e.WriteToken(json.ArrayEnd); err != nil {
		return err
	}
	return nil
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (n *Enum) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	offset := d.InputOffset()
	// Check Kind for invalid, next call will return error.
	if kind := d.PeekKind(); kind != '[' && kind != 0 {
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: d.StackPointer(),
			JSONKind:    kind,
			GoType:      reflect.TypeOf(n),
			Err:         errors.Errorf("unexpected type %s", kind.String()),
		}
	}
	// Read the opening bracket.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	// Keep non-nil value, to distinguish from not set array.
	values := Enum{}
	for {
		if kind := d.PeekKind(); kind == ']' || kind == 0 {
			break
		}
		var (
			val    json.RawValue
			offset = d.InputOffset()
		)
		if err := opts.UnmarshalNext(d, &val); err != nil {
			return err
		}
		for _, val2 := range values {
			if ok, _ := ogenjson.Equal(val, val2); ok {
				return &json.SemanticError{
					ByteOffset:  offset,
					JSONPointer: d.StackPointer(),
					JSONKind:    val.Kind(),
					GoType:      reflect.TypeOf(val),
					Err:         errors.Errorf("duplicate value %s", val.String()),
				}
			}
		}
		values = append(values, val)
	}
	// Read the closing bracket.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	*n = values
	return nil
}

// MarshalJSON implements json.MarshalerV1.
func (n *Enum) MarshalJSON() ([]byte, error) {
	// Backward-compatibility with v1.
	return json.Marshal(n)
}

// UnmarshalJSON implements json.UnmarshalerV1.
func (n *Enum) UnmarshalJSON(data []byte) error {
	// Backward-compatibility with v1.
	return json.Unmarshal(data, n)
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
