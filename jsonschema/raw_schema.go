package jsonschema

import (
	"bytes"
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	ogenjson "github.com/ogen-go/ogen/json"
)

// Num represents JSON number.
type Num jx.Num

// MarshalJSON implements json.Marshaler.
func (n Num) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.RawMessage(n))
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Num) UnmarshalJSON(data []byte) error {
	j, err := jx.DecodeBytes(data).Num()
	if err != nil {
		return errors.Wrapf(err, "invalid number %s", data)
	}
	if j.Str() {
		return errors.Errorf("invalid number %s", data)
	}

	*n = Num(j)
	return nil
}

// RawSchema is unparsed JSON Schema.
type RawSchema struct {
	Ref                  string                `json:"$ref,omitempty"`
	Summary              string                `json:"summary,omitempty"`
	Description          string                `json:"description,omitempty"`
	Type                 string                `json:"type,omitempty"`
	Format               string                `json:"format,omitempty"`
	Properties           RawProperties         `json:"properties,omitempty"`
	AdditionalProperties *AdditionalProperties `json:"additionalProperties,omitempty"`
	PatternProperties    RawPatternProperties  `json:"patternProperties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	Items                *RawSchema            `json:"items,omitempty"`
	Nullable             bool                  `json:"nullable,omitempty"`
	AllOf                []*RawSchema          `json:"allOf,omitempty"`
	OneOf                []*RawSchema          `json:"oneOf,omitempty"`
	AnyOf                []*RawSchema          `json:"anyOf,omitempty"`
	Discriminator        *Discriminator        `json:"discriminator,omitempty"`
	Enum                 Enum                  `json:"enum,omitempty"`
	MultipleOf           Num                   `json:"multipleOf,omitempty"`
	Maximum              Num                   `json:"maximum,omitempty"`
	ExclusiveMaximum     bool                  `json:"exclusiveMaximum,omitempty"`
	Minimum              Num                   `json:"minimum,omitempty"`
	ExclusiveMinimum     bool                  `json:"exclusiveMinimum,omitempty"`
	MaxLength            *uint64               `json:"maxLength,omitempty"`
	MinLength            *uint64               `json:"minLength,omitempty"`
	Pattern              string                `json:"pattern,omitempty"`
	MaxItems             *uint64               `json:"maxItems,omitempty"`
	MinItems             *uint64               `json:"minItems,omitempty"`
	UniqueItems          bool                  `json:"uniqueItems,omitempty"`
	MaxProperties        *uint64               `json:"maxProperties,omitempty"`
	MinProperties        *uint64               `json:"minProperties,omitempty"`
	Default              json.RawMessage       `json:"default,omitempty"`
	Example              json.RawMessage       `json:"example,omitempty"`
	Deprecated           bool                  `json:"deprecated,omitempty"`
	ContentEncoding      string                `json:"contentEncoding,omitempty"`
	ContentMediaType     string                `json:"contentMediaType,omitempty"`
	XAnnotations         map[string]jx.Raw     `json:"-"`
}

var xPrefix = []byte("x-")

// UnmarshalJSON implements json.Unmarshaler.
func (r *RawSchema) UnmarshalJSON(data []byte) error {
	type Alias RawSchema
	var val Alias
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	*r = RawSchema(val)

	d := jx.DecodeBytes(data)
	return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		if !bytes.HasPrefix(key, xPrefix) {
			return d.Skip()
		}
		val, err := d.RawAppend(nil)
		if err != nil {
			return errors.Wrapf(err, "decode %q", key)
		}
		if r.XAnnotations == nil {
			r.XAnnotations = map[string]jx.Raw{}
		}
		r.XAnnotations[string(key)] = val
		return nil
	})
}

// Enum is JSON Schema enum validator description.
type Enum []json.RawMessage

// UnmarshalJSON implements json.Unmarshaler.
func (n *Enum) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return d.Arr(func(d *jx.Decoder) error {
		val, err := d.RawAppend(nil)
		if err != nil {
			return errors.Wrapf(err, "parse [%d]", len(*n))
		}
		for _, x := range *n {
			if eq, _ := ogenjson.Equal(x, val); eq {
				return errors.Errorf("value %q is duplicated", x)
			}
		}
		*n = append(*n, json.RawMessage(val))
		return nil
	})
}

// RawProperty is item of RawProperties.
type RawProperty struct {
	Name   string
	Schema *RawSchema
}

// RawProperties is unparsed JSON Schema properties validator description.
type RawProperties []RawProperty

// MarshalJSON implements json.Marshaler.
func (p RawProperties) MarshalJSON() ([]byte, error) {
	var e jx.Encoder
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

// MarshalJSON implements json.Marshaler.
func (r RawPatternProperties) MarshalJSON() ([]byte, error) {
	var e jx.Encoder
	e.ObjStart()
	for _, prop := range r {
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
func (r *RawPatternProperties) UnmarshalJSON(data []byte) error {
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

		*r = append(*r, RawPatternProperty{
			Pattern: key,
			Schema:  s,
		})
		return nil
	})
}

// Discriminator is JSON Schema discriminator description.
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}
