package jsonschema

import (
	"regexp"

	ogenjson "github.com/ogen-go/ogen/json"
)

// SchemaType is a JSON Schema type.
type SchemaType string

const (
	// Empty is empty (unset) schema type.
	Empty SchemaType = "" // OneOf, AnyOf, AllOf.
	// Object is "object" schema type.
	Object SchemaType = "object"
	// Array is "array" schema type.
	Array SchemaType = "array"
	// Integer is "integer" schema type.
	Integer SchemaType = "integer"
	// Number is "number" schema type.
	Number SchemaType = "number"
	// String is "string" schema type.
	String SchemaType = "string"
	// Boolean is "boolean" schema type.
	Boolean SchemaType = "boolean"
	// Null is "null" schema type.
	Null SchemaType = "null"
)

// Schema is a JSON Schema.
type Schema struct {
	XOgenName string // Annotation to set type name.

	Ref string // Whether schema is referenced.

	Type             SchemaType
	Format           string // Schema format, optional.
	ContentEncoding  string
	ContentMediaType string

	Summary     string // Schema summary from Reference Object, optional.
	Description string // Schema description, optional.
	Deprecated  bool

	Item                 *Schema           // Only for Array and Object with additional properties.
	AdditionalProperties *bool             // Whether Object has additional properties.
	PatternProperties    []PatternProperty // Only for Object.
	Enum                 []interface{}     // Only for Enum.
	Properties           []Property        // Only for Object.

	Nullable bool // Whether schema is nullable or not. Any types.

	OneOf         []*Schema
	AnyOf         []*Schema
	AllOf         []*Schema
	Discriminator *Discriminator

	// Numeric validation (Integer, Number).
	Maximum          Num
	ExclusiveMaximum bool
	Minimum          Num
	ExclusiveMinimum bool
	MultipleOf       Num

	// String validation.
	MaxLength *uint64
	MinLength *uint64
	Pattern   string

	// Array validation.
	MaxItems    *uint64
	MinItems    *uint64
	UniqueItems bool

	// Object validation.
	MaxProperties *uint64
	MinProperties *uint64

	Examples []Example
	// Default schema value.
	Default    interface{}
	DefaultSet bool

	ogenjson.Locator `json:"-" yaml:"-"`
}

// AddExample adds example for this Schema.
func (s *Schema) AddExample(r Example) {
	if s != nil && len(r) > 0 {
		s.Examples = append(s.Examples, r)
	}
}

// Property is a JSON Schema Object property.
type Property struct {
	Name        string  // Property name.
	Description string  // Property description.
	Schema      *Schema // Property schema.
	Required    bool    // Whether the field is required or not.
}

// PatternProperty is a property pattern.
type PatternProperty struct {
	Pattern *regexp.Regexp
	Schema  *Schema
}
