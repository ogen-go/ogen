package jsonschema

import (
	"reflect"
	"slices"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/go-faster/yaml"
)

// RawType is the value of the "type" field of a JSON Schema.
//
// JSON Schema (and therefore OpenAPI 3.1) permits the value to be either a string or a list of strings.
// We represent a schema as a single type with a separate nullability flag, so a list like `[string, 'null']` is
// collapsed to Type "string" with Nullable set.
type RawType struct {
	// Type is a single schema type.
	Type string
	// Nullable reports whether the type list additionally contained "null".
	Nullable bool
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *RawType) UnmarshalJSON(data []byte) error {
	*t = RawType{}

	d := jx.DecodeBytes(data)
	switch tt := d.Next(); tt {
	case jx.Null:
		return d.Null()
	case jx.String:
		v, err := d.Str()
		if err != nil {
			return err
		}
		t.Type = v
		return nil
	case jx.Array:
		var types []string
		if err := d.Arr(func(d *jx.Decoder) error {
			if d.Next() == jx.Null {
				types = append(types, "null")
				return d.Null()
			}
			v, err := d.Str()
			if err != nil {
				return err
			}
			types = append(types, v)
			return nil
		}); err != nil {
			return err
		}

		typ, nullable, err := collapseTypeList(types)
		if err != nil {
			return err
		}
		t.Type, t.Nullable = typ, nullable
		return nil
	default:
		return errors.Errorf("unexpected type %s", tt.String())
	}
}

// CollapseTypeNode replaces a type list like `type: [string, 'null']` in the given mapping node with the
// single type it denotes and reports whether "null" was in the list.
//
// The result is a shallow copy. node itself is left untouched, since the same document may be decoded again
// by reference resolvers.
func CollapseTypeNode(node *yaml.Node) (_ *yaml.Node, nullable bool, err error) {
	if node.Kind != yaml.MappingNode {
		return node, false, nil
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		key, val := node.Content[i], node.Content[i+1]
		if val.Kind == yaml.AliasNode && val.Alias != nil {
			val = val.Alias
		}
		if key.Value != "type" || val.Kind != yaml.SequenceNode {
			continue
		}

		types := make([]string, len(val.Content))
		for j, item := range val.Content {
			// Tolerate an unquoted null.
			if item.ShortTag() == "!!null" {
				types[j] = "null"
				continue
			}
			if err := item.Decode(&types[j]); err != nil {
				return nil, false, err
			}
		}

		typ, nullable, err := collapseTypeList(types)
		if err != nil {
			return nil, false, &yaml.UnmarshalError{
				Node: val,
				Type: reflect.TypeOf(typ),
				Err:  err,
			}
		}

		copied := *node
		copied.Content = slices.Clone(node.Content)
		copied.Content[i+1] = &yaml.Node{
			Kind:   yaml.ScalarNode,
			Tag:    "!!str",
			Value:  typ,
			Line:   val.Line,
			Column: val.Column,
		}
		return &copied, nullable, nil
	}
	return node, false, nil
}

// collapseTypeList collapses a JSON Schema type list to a single type and a nullability flag.
func collapseTypeList(types []string) (typ string, nullable bool, _ error) {
	var rest []string
	for _, t := range types {
		if t == "null" {
			nullable = true
			continue
		}
		if !slices.Contains(rest, t) {
			rest = append(rest, t)
		}
	}
	switch len(rest) {
	case 0:
		if !nullable {
			return "", false, errors.New("type list must not be empty")
		}
		return "null", false, nil
	case 1:
		return rest[0], nullable, nil
	default:
		return "", false, errors.Errorf("multiple types are not supported: [%s]", strings.Join(types, ", "))
	}
}
