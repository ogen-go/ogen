package jsonschema

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/go-faster/yaml"
)

// Const is JSON Schema const validator description.
type Const json.RawMessage

// MarshalYAML implements yaml.Marshaler.
func (c Const) MarshalYAML() (any, error) {
	return convertJSONToRawYAML(json.RawMessage(c))
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Const) UnmarshalYAML(node *yaml.Node) error {
	raw, err := convertYAMLtoRawJSON(node)
	if err != nil {
		return &yaml.UnmarshalError{
			Node: node,
			Err:  errors.Wrapf(err, "cannot unmarshal %s into %T", node.ShortTag(), c),
		}
	}
	if err := validateConst(raw); err != nil {
		return &yaml.UnmarshalError{
			Node: node,
			Err:  err,
		}
	}
	*c = Const(raw)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (c Const) MarshalJSON() ([]byte, error) {
	e := &jx.Encoder{}
	e.Raw([]byte(c))
	return e.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Const) UnmarshalJSON(b []byte) error {
	raw, err := jx.DecodeBytes(b).Raw()
	if err != nil {
		return err
	}
	if err := validateConst(raw); err != nil {
		return err
	}
	*c = Const(raw)
	return nil
}

// validateConst rejects empty objects.
func validateConst(raw []byte) error {
	d := jx.DecodeBytes(raw)
	if d.Next() != jx.Object {
		return nil
	}
	empty := true
	if err := d.Obj(func(d *jx.Decoder, key string) error {
		empty = false
		return d.Skip()
	}); err != nil {
		return err
	}
	if empty {
		return errors.New("const cannot be an empty object")
	}
	return nil
}
