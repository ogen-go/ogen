package jsonschema

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
	Items                *RawItems             `json:"items,omitempty" yaml:"items,omitempty"`
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

	Discriminator *RawDiscriminator `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	XML           *XML              `json:"xml,omitempty" yaml:"xml,omitempty"`
	Example       Example           `json:"example,omitempty" yaml:"example,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// RawDiscriminator discriminates types for OneOf, AllOf, AnyOf.
//
// See https://spec.openapis.org/oas/v3.1.0#discriminator-object.
type RawDiscriminator struct {
	// REQUIRED. The name of the property in the payload that will hold the discriminator value.
	PropertyName string `json:"propertyName" yaml:"propertyName"`
	// An object to hold mappings between payload values and schema names or references.
	Mapping map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
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
