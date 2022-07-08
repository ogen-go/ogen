package ogen

import (
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"
	"gopkg.in/yaml.v3"

	ogenjson "github.com/ogen-go/ogen/json"
)

// The Schema Object allows the definition of input and output data types.
// These types can be objects, but also primitives and arrays.
type Schema struct {
	Ref         string `json:"$ref,omitempty,omitzero" yaml:"$ref,omitempty"` // ref object
	Summary     string `json:"summary,omitempty,omitzero" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty,omitzero" yaml:"description,omitempty"`

	// Additional external documentation for this schema.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty,omitzero" yaml:"externalDocs,omitempty"`

	// Value MUST be a string. Multiple types via an array are not supported.
	Type string `json:"type,omitempty,omitzero" yaml:"type,omitempty"`

	// See Data Type Formats for further details (https://swagger.io/specification/#data-type-format).
	// While relying on JSON Schema's defined formats,
	// the OAS offers a few additional predefined formats.
	Format string `json:"format,omitempty,omitzero" yaml:"format,omitempty"`

	// Property definitions MUST be a Schema Object and not a standard JSON Schema
	// (inline or referenced).
	Properties Properties `json:"properties,omitempty,omitzero" yaml:"properties,omitempty"`

	// Value can be boolean or object. Inline or referenced schema MUST be of a Schema Object
	// and not a standard JSON Schema. Consistent with JSON Schema, additionalProperties defaults to true.
	AdditionalProperties *AdditionalProperties `json:"additionalProperties,omitempty,omitzero" yaml:"additionalProperties,omitempty"`

	// The value of "patternProperties" MUST be an object. Each property
	// name of this object SHOULD be a valid regular expression, according
	// to the ECMA-262 regular expression dialect. Each property value of
	// this object MUST be a valid JSON Schema.
	PatternProperties PatternProperties `json:"patternProperties,omitempty,omitzero" yaml:"patternProperties,omitempty"`

	// The value of this keyword MUST be an array.
	// This array MUST have at least one element.
	// Elements of this array MUST be strings, and MUST be unique.
	Required []string `json:"required,omitempty,omitzero" yaml:"required,omitempty"`

	// Value MUST be an object and not an array.
	// Inline or referenced schema MUST be of a Schema Object and not a standard JSON Schema.
	// MUST be present if the Type is "array".
	Items *Schema `json:"items,omitempty,omitzero" yaml:"items,omitempty"`

	// A true value adds "null" to the allowed type specified by the type keyword,
	// only if type is explicitly defined within the same Schema Object.
	// Other Schema Object constraints retain their defined behavior,
	// and therefore may disallow the use of null as a value.
	// A false value leaves the specified or default type unmodified.
	// The default value is false.
	Nullable bool `json:"nullable,omitempty,omitzero" yaml:"nullable,omitempty"`

	// AllOf takes an array of object definitions that are used
	// for independent validation but together compose a single object.
	// Still, it does not imply a hierarchy between the models.
	// For that purpose, you should include the discriminator.
	AllOf []*Schema `json:"allOf,omitempty,omitzero" yaml:"allOf,omitempty"`

	// OneOf validates the value against exactly one of the subschemas
	OneOf []*Schema `json:"oneOf,omitempty,omitzero" yaml:"oneOf,omitempty"`

	// AnyOf validates the value against any (one or more) of the subschemas
	AnyOf []*Schema `json:"anyOf,omitempty,omitzero" yaml:"anyOf,omitempty"`

	// Discriminator for subschemas.
	Discriminator *Discriminator `json:"discriminator,omitempty,omitzero" yaml:"discriminator,omitempty"`

	// The value of this keyword MUST be an array.
	// This array SHOULD have at least one element.
	// Elements in the array SHOULD be unique.
	Enum Enum `json:"enum,omitempty,omitzero" yaml:"enum,omitempty"`

	// The value of "multipleOf" MUST be a number, strictly greater than 0.
	//
	// A numeric instance is only valid if division by this keyword's value
	// results in an integer.
	MultipleOf Num `json:"multipleOf,omitempty,omitzero" yaml:"multipleOf,omitempty"`

	// The value of "maximum" MUST be a number, representing an upper limit
	// for a numeric instance.
	//
	// If the instance is a number, then this keyword validates if
	// "exclusiveMaximum" is true and instance is less than the provided
	// value, or else if the instance is less than or exactly equal to the
	// provided value.
	Maximum Num `json:"maximum,omitempty,omitzero" yaml:"maximum,omitempty"`

	// The value of "exclusiveMaximum" MUST be a boolean, representing
	// whether the limit in "maximum" is exclusive or not.  An undefined
	// value is the same as false.
	//
	// If "exclusiveMaximum" is true, then a numeric instance SHOULD NOT be
	// equal to the value specified in "maximum".  If "exclusiveMaximum" is
	// false (or not specified), then a numeric instance MAY be equal to the
	// value of "maximum".
	ExclusiveMaximum bool `json:"exclusiveMaximum,omitempty,omitzero" yaml:"exclusiveMaximum,omitempty"`

	// The value of "minimum" MUST be a number, representing a lower limit
	// for a numeric instance.
	//
	// If the instance is a number, then this keyword validates if
	// "exclusiveMinimum" is true and instance is greater than the provided
	// value, or else if the instance is greater than or exactly equal to
	// the provided value.
	Minimum Num `json:"minimum,omitempty,omitzero" yaml:"minimum,omitempty"`

	// The value of "exclusiveMinimum" MUST be a boolean, representing
	// whether the limit in "minimum" is exclusive or not.  An undefined
	// value is the same as false.
	//
	// If "exclusiveMinimum" is true, then a numeric instance SHOULD NOT be
	// equal to the value specified in "minimum".  If "exclusiveMinimum" is
	// false (or not specified), then a numeric instance MAY be equal to the
	// value of "minimum".
	ExclusiveMinimum bool `json:"exclusiveMinimum,omitempty,omitzero" yaml:"exclusiveMinimum,omitempty"`

	// The value of this keyword MUST be a non-negative integer.
	//
	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// A string instance is valid against this keyword if its length is less
	// than, or equal to, the value of this keyword.
	//
	// The length of a string instance is defined as the number of its
	// characters as defined by RFC 7159 [RFC7159].
	MaxLength *uint64 `json:"maxLength,omitempty,omitzero" yaml:"maxLength,omitempty"`

	// A string instance is valid against this keyword if its length is
	// greater than, or equal to, the value of this keyword.
	//
	// The length of a string instance is defined as the number of its
	// characters as defined by RFC 7159 [RFC7159].
	//
	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// "minLength", if absent, may be considered as being present with
	// integer value 0.
	MinLength *uint64 `json:"minLength,omitempty,omitzero" yaml:"minLength,omitempty"`

	// The value of this keyword MUST be a string.  This string SHOULD be a
	// valid regular expression, according to the ECMA 262 regular
	// expression dialect.
	//
	// A string instance is considered valid if the regular expression
	// matches the instance successfully. Recall: regular expressions are
	// not implicitly anchored.
	Pattern string `json:"pattern,omitempty,omitzero" yaml:"pattern,omitempty"`

	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// An array instance is valid against "maxItems" if its size is less
	// than, or equal to, the value of this keyword.
	MaxItems *uint64 `json:"maxItems,omitempty,omitzero" yaml:"maxItems,omitempty"`

	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// An array instance is valid against "minItems" if its size is greater
	// than, or equal to, the value of this keyword.
	//
	// If this keyword is not present, it may be considered present with a
	// value of 0.
	MinItems *uint64 `json:"minItems,omitempty,omitzero" yaml:"minItems,omitempty"`

	// The value of this keyword MUST be a boolean.
	//
	// If this keyword has boolean value false, the instance validates
	// successfully.  If it has boolean value true, the instance validates
	// successfully if all of its elements are unique.
	//
	// If not present, this keyword may be considered present with boolean
	// value false.
	UniqueItems bool `json:"uniqueItems,omitempty,omitzero" yaml:"uniqueItems,omitempty"`

	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// An object instance is valid against "maxProperties" if its number of
	// properties is less than, or equal to, the value of this keyword.
	MaxProperties *uint64 `json:"maxProperties,omitempty,omitzero" yaml:"maxProperties,omitempty"`

	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// An object instance is valid against "minProperties" if its number of
	// properties is greater than, or equal to, the value of this keyword.
	//
	// If this keyword is not present, it may be considered present with a
	// value of 0.
	MinProperties *uint64 `json:"minProperties,omitempty,omitzero" yaml:"minProperties,omitempty"`

	// Default value.
	Default Default `json:"default,omitempty,omitzero" yaml:"default,omitempty"`

	// A free-form property to include an example of an instance for this schema.
	// To represent examples that cannot be naturally represented in JSON or YAML,
	// a string value can be used to contain the example with escaping where necessary.
	Example ExampleValue `json:"example,omitempty,omitzero" yaml:"example,omitempty"`

	// Specifies that a schema is deprecated and SHOULD be transitioned out
	// of usage.
	Deprecated bool `json:"deprecated,omitempty,omitzero" yaml:"deprecated,omitempty"`

	// If the instance value is a string, this property defines that the
	// string SHOULD be interpreted as binary data and decoded using the
	// encoding named by this property.  RFC 2045, Section 6.1 lists
	// the possible values for this property.
	//
	// The value of this property MUST be a string.
	//
	// The value of this property SHOULD be ignored if the instance
	// described is not a string.
	ContentEncoding string `json:"contentEncoding,omitempty,omitzero" yaml:"contentEncoding,omitempty"`

	// The value of this property must be a media type, as defined by RFC
	// 2046. This property defines the media type of instances
	// which this schema defines.
	//
	// The value of this property MUST be a string.
	//
	// The value of this property SHOULD be ignored if the instance
	// described is not a string.
	ContentMediaType string `json:"contentMediaType,omitempty,omitzero" yaml:"contentMediaType,omitempty"`

	Locator `json:"-" yaml:"-"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (s *Schema) UnmarshalYAML(n *yaml.Node) error {
	type Alias Schema
	var a Alias

	if err := n.Decode(&a); err != nil {
		return err
	}
	a.SetLocation(ogenjson.Location{
		// FIXME(tdakkota): For some reason, the line number is off by 1.
		//  And the column number is off by 2.
		Line:   int64(n.Line) - 1,
		Column: int64(n.Column) - 2,
	})

	*s = Schema(a)
	return nil
}

// Property is item of Properties.
type Property struct {
	Name   string
	Schema *Schema
}

// Properties is unparsed JSON Schema properties validator description.
type Properties []Property

// MarshalNextJSON implements json.MarshalerV2.
func (p Properties) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
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
func (p *Properties) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
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
	properties := Properties{}
	for d.PeekKind() == '"' {
		var (
			name   string
			schema *Schema
		)
		if err := opts.UnmarshalNext(d, &name); err != nil {
			return err
		}
		if err := opts.UnmarshalNext(d, &schema); err != nil {
			return err
		}
		properties = append(properties, Property{Name: name, Schema: schema})
	}
	// Read the closing brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	*p = properties
	return nil
}

// MarshalJSON implements json.MarshalerV1.
func (p *Properties) MarshalJSON() ([]byte, error) {
	// Backward-compatibility with v1.
	return json.Marshal(p)
}

// UnmarshalJSON implements json.UnmarshalerV1.
func (p *Properties) UnmarshalJSON(data []byte) error {
	// Backward-compatibility with v1.
	return json.Unmarshal(data, p)
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *Properties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.Errorf("unexpected YAML kind %v", node.Kind)
	}
	for i := 0; i < len(node.Content); i += 2 {
		var (
			key    = node.Content[i]
			value  = node.Content[i+1]
			schema *Schema
		)
		if err := value.Decode(&schema); err != nil {
			return err
		}
		*p = append(*p, Property{
			Name:   key.Value,
			Schema: schema,
		})
	}
	return nil
}

// AdditionalProperties is JSON Schema additionalProperties validator description.
type AdditionalProperties struct {
	Bool   *bool
	Schema Schema
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

// MarshalJSON implements json.MarshalerV1.
func (p *AdditionalProperties) MarshalJSON() ([]byte, error) {
	// Backward-compatibility with v1.
	return json.Marshal(p)
}

// UnmarshalJSON implements json.UnmarshalerV1.
func (p *AdditionalProperties) UnmarshalJSON(data []byte) error {
	// Backward-compatibility with v1.
	return json.Unmarshal(data, p)
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

// PatternProperty is item of PatternProperties.
type PatternProperty struct {
	Pattern string
	Schema  *Schema
}

// PatternProperties is unparsed JSON Schema patternProperties validator description.
type PatternProperties []PatternProperty

// MarshalNextJSON implements json.MarshalerV2.
func (p PatternProperties) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	if err := e.WriteToken(json.ObjectStart); err != nil {
		return err
	}
	for _, member := range p {
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
func (p *PatternProperties) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
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
	patternProperties := PatternProperties{}
	for d.PeekKind() == '"' {
		var (
			pattern string
			schema  *Schema
		)
		if err := opts.UnmarshalNext(d, &pattern); err != nil {
			return err
		}
		if err := opts.UnmarshalNext(d, &schema); err != nil {
			return err
		}
		patternProperties = append(patternProperties, PatternProperty{Pattern: pattern, Schema: schema})
	}
	// Read the closing brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	*p = patternProperties
	return nil
}

// MarshalJSON implements json.MarshalerV1.
func (p *PatternProperties) MarshalJSON() ([]byte, error) {
	// Backward-compatibility with v1.
	return json.Marshal(p)
}

// UnmarshalJSON implements json.UnmarshalerV1.
func (p *PatternProperties) UnmarshalJSON(data []byte) error {
	// Backward-compatibility with v1.
	return json.Unmarshal(data, p)
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *PatternProperties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.Errorf("unexpected YAML kind %v", node.Kind)
	}
	for i := 0; i < len(node.Content); i += 2 {
		var (
			key    = node.Content[i]
			value  = node.Content[i+1]
			schema *Schema
		)
		if err := value.Decode(&schema); err != nil {
			return err
		}
		*p = append(*p, PatternProperty{
			Pattern: key.Value,
			Schema:  schema,
		})
	}
	return nil
}
