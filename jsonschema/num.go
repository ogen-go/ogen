package jsonschema

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	yaml "github.com/go-faster/yamlx"
)

// Num represents JSON number.
type Num json.RawMessage

// UnmarshalYAML implements yaml.Unmarshaler.
func (n *Num) UnmarshalYAML(node *yaml.Node) error {
	if t := node.Tag; t != "!!int" && t != "!!float" {
		return errors.Errorf("unexpected tag %s", t)
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
