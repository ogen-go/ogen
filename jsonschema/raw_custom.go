package jsonschema

import (
	"encoding/json"
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	yaml "github.com/go-faster/yamlx"
)

// RawProperty is item of RawProperties.
type RawProperty struct {
	Name   string
	Schema *RawSchema
}

// RawProperties is unparsed JSON Schema properties validator description.
type RawProperties []RawProperty

// MarshalYAML implements yaml.Marshaler.
func (p RawProperties) MarshalYAML() (interface{}, error) {
	content := make([]*yaml.Node, 0, len(p)*2)
	for _, prop := range p {
		var val yaml.Node
		if err := val.Encode(prop.Schema); err != nil {
			return nil, err
		}
		content = append(content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: prop.Name},
			&val,
		)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: content,
	}, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *RawProperties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return &yaml.UnmarshalError{
			Node: node,
			Type: reflect.TypeOf(p),
			Err:  errors.Errorf("cannot unmarshal %s into %T", node.ShortTag(), p),
		}
	}
	for i := 0; i < len(node.Content); i += 2 {
		var (
			key    = node.Content[i]
			value  = node.Content[i+1]
			schema *RawSchema
		)
		if err := value.Decode(&schema); err != nil {
			return err
		}
		*p = append(*p, RawProperty{
			Name:   key.Value,
			Schema: schema,
		})
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (p RawProperties) MarshalJSON() ([]byte, error) {
	e := jx.GetEncoder()
	defer jx.PutEncoder(e)

	e.ObjStart()
	for _, prop := range p {
		e.FieldStart(prop.Name)
		b, err := json.Marshal(prop.Schema)
		if err != nil {
			return nil, errors.Wrap(err, "marshal")
		}
		e.Raw(b)
	}
	e.ObjEnd()
	return e.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *RawProperties) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return d.Obj(func(d *jx.Decoder, key string) error {
		s := new(RawSchema)
		b, err := d.Raw()
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, s); err != nil {
			return err
		}

		*p = append(*p, RawProperty{
			Name:   key,
			Schema: s,
		})
		return nil
	})
}

// AdditionalProperties is JSON Schema additionalProperties validator description.
type AdditionalProperties struct {
	Bool   *bool
	Schema RawSchema
}

// MarshalYAML implements yaml.Marshaler.
func (p AdditionalProperties) MarshalYAML() (interface{}, error) {
	if p.Bool != nil {
		return *p.Bool, nil
	}
	return p.Schema, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *AdditionalProperties) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		return node.Decode(&p.Bool)
	case yaml.MappingNode:
		return node.Decode(&p.Schema)
	default:
		return errors.Errorf("unexpected YAML kind %v", node.Kind)
	}
}

// MarshalJSON implements json.Marshaler.
func (p AdditionalProperties) MarshalJSON() ([]byte, error) {
	if p.Bool != nil {
		return json.Marshal(p.Bool)
	}
	return json.Marshal(p.Schema)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *AdditionalProperties) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	switch tt := d.Next(); tt {
	case jx.Object:
	case jx.Bool:
		val, err := d.Bool()
		if err != nil {
			return err
		}
		p.Bool = &val
		return nil
	default:
		return errors.Errorf("unexpected type %s", tt.String())
	}

	s := RawSchema{}
	b, err := d.Raw()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	p.Schema = s
	return nil
}

// RawPatternProperty is item of RawPatternProperties.
type RawPatternProperty struct {
	Pattern string
	Schema  *RawSchema
}

// RawPatternProperties is unparsed JSON Schema patternProperties validator description.
type RawPatternProperties []RawPatternProperty

// MarshalYAML implements yaml.Marshaler.
func (p RawPatternProperties) MarshalYAML() (interface{}, error) {
	content := make([]*yaml.Node, 0, len(p)*2)
	for _, prop := range p {
		var val yaml.Node
		if err := val.Encode(prop.Schema); err != nil {
			return nil, err
		}
		content = append(content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: prop.Pattern},
			&val,
		)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: content,
	}, nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *RawPatternProperties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return &yaml.UnmarshalError{
			Node: node,
			Type: reflect.TypeOf(p),
			Err:  errors.Errorf("cannot unmarshal %s into %T", node.ShortTag(), p),
		}
	}
	for i := 0; i < len(node.Content); i += 2 {
		var (
			key    = node.Content[i]
			value  = node.Content[i+1]
			schema *RawSchema
		)
		if err := value.Decode(&schema); err != nil {
			return err
		}
		*p = append(*p, RawPatternProperty{
			Pattern: key.Value,
			Schema:  schema,
		})
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (p RawPatternProperties) MarshalJSON() ([]byte, error) {
	var e jx.Encoder
	e.ObjStart()
	for _, prop := range p {
		e.FieldStart(prop.Pattern)
		b, err := json.Marshal(prop.Schema)
		if err != nil {
			return nil, errors.Wrap(err, "marshal")
		}
		e.Raw(b)
	}
	e.ObjEnd()
	return e.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *RawPatternProperties) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return d.Obj(func(d *jx.Decoder, key string) error {
		s := new(RawSchema)
		b, err := d.Raw()
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, s); err != nil {
			return err
		}

		*p = append(*p, RawPatternProperty{
			Pattern: key,
			Schema:  s,
		})
		return nil
	})
}
