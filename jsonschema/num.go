package jsonschema

import (
	"encoding/json"
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/go-faster/yaml"
)

// Num represents JSON number.
type Num json.RawMessage

// MarshalYAML implements yaml.Marshaler.
func (n Num) MarshalYAML() (any, error) {
	tag := "!!float"
	if jx.Num(n).IsInt() {
		tag = "!!int"
	}
	return yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   tag,
		Value: string(n),
	}, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (n *Num) UnmarshalYAML(node *yaml.Node) error {
	if t := node.Tag; node.Kind != yaml.ScalarNode || (t != "!!int" && t != "!!float") {
		return &yaml.UnmarshalError{
			Node: node,
			Type: reflect.TypeOf(n),
			Err:  errors.Errorf("cannot unmarshal %s into %T", node.ShortTag(), n),
		}
	}

	val, err := convertYAMLtoRawJSON(node)
	if err != nil {
		return err
	}

	num, err := jx.DecodeBytes(val).Num()
	if err != nil {
		return err
	}
	*n = Num(num)

	return nil
}

// MarshalJSON implements json.Marshaler.
func (n Num) MarshalJSON() ([]byte, error) {
	return json.RawMessage(n).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Num) UnmarshalJSON(data []byte) error {
	num, err := jx.DecodeBytes(data).Num()
	if err != nil {
		return err
	}
	if num.Str() {
		return errors.Errorf("unexpected string %s", num)
	}
	*n = Num(num)
	return nil
}
