package jsonschema

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/go-faster/yaml"

	"github.com/ogen-go/ogen/location"
)

// Extensions map is "^x-" fields list.
type Extensions map[string]yaml.Node

func isExtensionKey(key string) bool {
	return strings.HasPrefix(key, "x-")
}

// MarshalYAML implements yaml.Marshaler.
func (p Extensions) MarshalYAML() (any, error) {
	content := make([]*yaml.Node, 0, len(p)*2)
	for key, val := range p {
		val := val
		if !isExtensionKey(key) {
			continue
		}
		content = append(content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
			&val,
		)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: content,
	}, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *Extensions) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return &yaml.UnmarshalError{
			Node: node,
			Type: reflect.TypeOf(p),
			Err:  errors.Errorf("cannot unmarshal %s into %T", node.ShortTag(), p),
		}
	}
	m := *p
	if m == nil {
		m = make(Extensions, len(node.Content)/2)
		*p = m
	}
	for i := 0; i < len(node.Content); i += 2 {
		var (
			keyNode = node.Content[i]
			value   = node.Content[i+1]
			key     string
		)
		if err := keyNode.Decode(&key); err != nil {
			return err
		}
		// FIXME(tdakkota): use *yamlx.Node instead of yaml.Node
		if isExtensionKey(key) && value != nil {
			m[key] = *value
		}
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (p Extensions) MarshalJSON() ([]byte, error) {
	e := &jx.Encoder{}

	e.ObjStart()
	for key, val := range p {
		val := val
		if !isExtensionKey(key) {
			continue
		}

		e.FieldStart(key)
		b, err := convertYAMLtoRawJSON(&val)
		if err != nil {
			return nil, errors.Wrap(err, "marshal")
		}
		e.Raw(b)
	}
	e.ObjEnd()
	return e.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Extensions) UnmarshalJSON(data []byte) error {
	m := *p
	if m == nil {
		m = make(Extensions)
		*p = m
	}

	d := jx.DecodeBytes(data)
	return d.Obj(func(d *jx.Decoder, key string) error {
		if !isExtensionKey(key) {
			return d.Skip()
		}

		b, err := d.Raw()
		if err != nil {
			return err
		}

		value, err := convertJSONToRawYAML(json.RawMessage(b))
		if err != nil {
			return err
		}
		m[key] = *value

		return nil
	})
}

// OpenAPICommon is common fields for OpenAPI objects.
type OpenAPICommon struct {
	Extensions
	location.Locator
}

// MarshalYAML implements yaml.Marshaler.
func (p OpenAPICommon) MarshalYAML() (any, error) {
	return p.Extensions.MarshalYAML()
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *OpenAPICommon) UnmarshalYAML(node *yaml.Node) error {
	if err := p.Extensions.UnmarshalYAML(node); err != nil {
		return errors.Wrap(err, "unmarshal extensions")
	}
	if err := p.Locator.UnmarshalYAML(node); err != nil {
		return errors.Wrap(err, "unmarshal locator")
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (p OpenAPICommon) MarshalJSON() ([]byte, error) {
	return p.Extensions.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *OpenAPICommon) UnmarshalJSON(data []byte) error {
	return p.Extensions.UnmarshalJSON(data)
}
