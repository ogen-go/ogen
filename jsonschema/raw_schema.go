package jsonschema

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	yaml "github.com/go-faster/yamlx"

	"github.com/ogen-go/ogen/internal/location"
)

// RawSchema is unparsed JSON Schema.
type RawSchema struct {
	Ref                  string                `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary              string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description          string                `json:"description,omitempty" yaml:"description,omitempty"`
	Type                 string                `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string                `json:"format,omitempty" yaml:"format,omitempty"`
	Properties           RawProperties         `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties *AdditionalProperties `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	PatternProperties    RawPatternProperties  `json:"patternProperties,omitempty" yaml:"patternProperties,omitempty"`
	Required             []string              `json:"required,omitempty" yaml:"required,omitempty"`
	Items                *RawSchema            `json:"items,omitempty" yaml:"items,omitempty"`
	Nullable             bool                  `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	AllOf                []*RawSchema          `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []*RawSchema          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*RawSchema          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Enum                 Enum                  `json:"enum,omitempty" yaml:"enum,omitempty"`
	MultipleOf           Num                   `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum              Num                   `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum     bool                  `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum              Num                   `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum     bool                  `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength            *uint64               `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength            *uint64               `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern              string                `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems             *uint64               `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems             *uint64               `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems          bool                  `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties        *uint64               `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        *uint64               `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Default              Default               `json:"default,omitempty" yaml:"default,omitempty"`
	Deprecated           bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	ContentEncoding      string                `json:"contentEncoding,omitempty" yaml:"contentEncoding,omitempty"`
	ContentMediaType     string                `json:"contentMediaType,omitempty" yaml:"contentMediaType,omitempty"`

	Discriminator *Discriminator `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	XML           *XML           `json:"xml,omitempty" yaml:"xml,omitempty"`
	Example       Example        `json:"example,omitempty" yaml:"example,omitempty"`

	XAnnotations map[string]json.RawMessage `json:"-" yaml:"-"`
	Locator      location.Locator           `json:"-" yaml:",inline"`
}

// RawProperty is item of RawProperties.
type RawProperty struct {
	Name   string
	Schema *RawSchema
}

// RawProperties is unparsed JSON Schema properties validator description.
type RawProperties []RawProperty

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *RawProperties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.Errorf("unexpected YAML kind %v", node.Kind)
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

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *RawPatternProperties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.Errorf("unexpected YAML kind %v", node.Kind)
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

// Discriminator discriminates types for OneOf, AllOf, AnyOf.
//
// See https://spec.openapis.org/oas/v3.1.0#discriminator-object.
type Discriminator struct {
	// REQUIRED. The name of the property in the payload that will hold the discriminator value.
	PropertyName string `json:"propertyName" yaml:"propertyName"`
	// An object to hold mappings between payload values and schema names or references.
	Mapping map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

// XML is a metadata object that allows for more fine-tuned XML model definitions.
//
// See https://spec.openapis.org/oas/v3.1.0#xml-object.
type XML struct {
	// Replaces the name of the element/attribute used for the described schema property.
	//
	// When defined within items, it will affect the name of the individual XML elements within the list.
	//
	// When defined alongside type being array (outside the items), it will affect the wrapping element
	// and only if wrapped is true.
	//
	// If wrapped is false, it will be ignored.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// The URI of the namespace definition.
	//
	// This MUST be in the form of an absolute URI.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// The prefix to be used for the name.
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	// Declares whether the property definition translates to an attribute instead of an element.
	//
	// Default value is false.
	Attribute bool `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	// MAY be used only for an array definition. Signifies whether the array is wrapped
	// (for example, `<books><book/><book/></books>`) or unwrapped (`<book/><book/>`).
	//
	// The definition takes effect only when defined alongside type being array (outside the items).
	//
	// Default value is false.
	Wrapped bool `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
}
