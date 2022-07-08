package jsonschema

import (
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"

	ogenjson "github.com/ogen-go/ogen/json"
)

// RawSchema is unparsed JSON Schema.
type RawSchema struct {
	Ref                  string                `json:"$ref,omitempty,omitzero"`
	Summary              string                `json:"summary,omitempty,omitzero"`
	Description          string                `json:"description,omitempty,omitzero"`
	Type                 string                `json:"type,omitempty,omitzero"`
	Format               string                `json:"format,omitempty,omitzero"`
	Properties           RawProperties         `json:"properties,omitempty,omitzero"`
	AdditionalProperties *AdditionalProperties `json:"additionalProperties,omitempty,omitzero"`
	PatternProperties    RawPatternProperties  `json:"patternProperties,omitempty,omitzero"`
	Required             []string              `json:"required,omitempty,omitzero"`
	Items                *RawSchema            `json:"items,omitempty,omitzero"`
	Nullable             bool                  `json:"nullable,omitempty,omitzero"`
	AllOf                []*RawSchema          `json:"allOf,omitempty,omitzero"`
	OneOf                []*RawSchema          `json:"oneOf,omitempty,omitzero"`
	AnyOf                []*RawSchema          `json:"anyOf,omitempty,omitzero"`
	Discriminator        *Discriminator        `json:"discriminator,omitempty,omitzero"`
	Enum                 Enum                  `json:"enum,omitempty,omitzero"`
	MultipleOf           Num                   `json:"multipleOf,omitempty,omitzero"`
	Maximum              Num                   `json:"maximum,omitempty,omitzero"`
	ExclusiveMaximum     bool                  `json:"exclusiveMaximum,omitempty,omitzero"`
	Minimum              Num                   `json:"minimum,omitempty,omitzero"`
	ExclusiveMinimum     bool                  `json:"exclusiveMinimum,omitempty,omitzero"`
	MaxLength            *uint64               `json:"maxLength,omitempty,omitzero"`
	MinLength            *uint64               `json:"minLength,omitempty,omitzero"`
	Pattern              string                `json:"pattern,omitempty,omitzero"`
	MaxItems             *uint64               `json:"maxItems,omitempty,omitzero"`
	MinItems             *uint64               `json:"minItems,omitempty,omitzero"`
	UniqueItems          bool                  `json:"uniqueItems,omitempty,omitzero"`
	MaxProperties        *uint64               `json:"maxProperties,omitempty,omitzero"`
	MinProperties        *uint64               `json:"minProperties,omitempty,omitzero"`
	Default              Default               `json:"default,omitempty,omitzero"`
	Example              Example               `json:"example,omitempty,omitzero"`
	Deprecated           bool                  `json:"deprecated,omitempty,omitzero"`
	ContentEncoding      string                `json:"contentEncoding,omitempty,omitzero"`
	ContentMediaType     string                `json:"contentMediaType,omitempty,omitzero"`

	XAnnotations map[string]json.RawValue `json:",inline"`

	ogenjson.Locator `json:"-"`
}

// RawProperty is item of RawProperties.
type RawProperty struct {
	Name   string
	Schema *RawSchema
}

// RawProperties is unparsed JSON Schema properties validator description.
type RawProperties []RawProperty

// MarshalNextJSON implements json.MarshalerV2.
func (p RawProperties) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	if err := e.WriteToken(json.ObjectStart); err != nil {
		return err
	}
	for _, member := range p {
		if err := opts.MarshalNext(e, member.Name); err != nil {
			return err
		}
		if err := opts.MarshalNext(e, member.Schema); err != nil {
			return err
		}
	}
	if err := e.WriteToken(json.ObjectEnd); err != nil {
		return err
	}
	return nil
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (p *RawProperties) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	offset := d.InputOffset()
	// Check Kind for invalid, next call will return error.
	if kind := d.PeekKind(); kind != '{' && kind != 0 {
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: d.StackPointer(),
			JSONKind:    kind,
			GoType:      reflect.TypeOf(p),
			Err:         errors.Errorf("unexpected type %s", kind.String()),
		}
	}
	// Read the opening brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	// Keep non-nil value, to distinguish from not set object.
	properties := RawProperties{}
	for d.PeekKind() == '"' {
		var (
			name   string
			schema *RawSchema
		)
		if err := opts.UnmarshalNext(d, &name); err != nil {
			return err
		}
		if err := opts.UnmarshalNext(d, &schema); err != nil {
			return err
		}
		properties = append(properties, RawProperty{Name: name, Schema: schema})
	}
	// Read the closing brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	*p = properties
	return nil
}

// AdditionalProperties is JSON Schema additionalProperties validator description.
type AdditionalProperties struct {
	Bool   *bool
	Schema RawSchema
}

// MarshalNextJSON implements json.MarshalerV2.
func (p AdditionalProperties) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	if p.Bool != nil {
		return opts.MarshalNext(e, p.Bool)
	}
	return opts.MarshalNext(e, p.Schema)
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (p *AdditionalProperties) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	offset := d.InputOffset()
	switch kind := d.PeekKind(); kind {
	case 't', 'f':
		return opts.UnmarshalNext(d, &p.Bool)
	case '{':
		return opts.UnmarshalNext(d, &p.Schema)
	default:
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: d.StackPointer(),
			JSONKind:    kind,
			GoType:      reflect.TypeOf(p),
			Err:         errors.Errorf("unexpected type %s", kind.String()),
		}
	}
}

// RawPatternProperty is item of RawPatternProperties.
type RawPatternProperty struct {
	Pattern string
	Schema  *RawSchema
}

// RawPatternProperties is unparsed JSON Schema patternProperties validator description.
type RawPatternProperties []RawPatternProperty

// MarshalNextJSON implements json.MarshalerV2.
func (r RawPatternProperties) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	if err := e.WriteToken(json.ObjectStart); err != nil {
		return err
	}
	for _, member := range r {
		if err := opts.MarshalNext(e, member.Pattern); err != nil {
			return err
		}
		if err := opts.MarshalNext(e, member.Schema); err != nil {
			return err
		}
	}
	if err := e.WriteToken(json.ObjectEnd); err != nil {
		return err
	}
	return nil
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (r *RawPatternProperties) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	offset := d.InputOffset()
	// Check Kind for invalid, next call will return error.
	if kind := d.PeekKind(); kind != '{' && kind != 0 {
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: d.StackPointer(),
			JSONKind:    kind,
			GoType:      reflect.TypeOf(r),
			Err:         errors.Errorf("unexpected type %s", kind.String()),
		}
	}
	// Read the opening brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	// Keep non-nil value, to distinguish from not set object.
	patternProperties := RawPatternProperties{}
	for d.PeekKind() == '"' {
		var (
			pattern string
			schema  *RawSchema
		)
		if err := opts.UnmarshalNext(d, &pattern); err != nil {
			return err
		}
		if err := opts.UnmarshalNext(d, &schema); err != nil {
			return err
		}
		patternProperties = append(patternProperties, RawPatternProperty{Pattern: pattern, Schema: schema})
	}
	// Read the closing brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	*r = patternProperties
	return nil
}

// Discriminator is JSON Schema discriminator description.
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty,omitzero"`
}
