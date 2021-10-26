package ast

// SchemaType is an OpenAPI JSON Schema type.
type SchemaType string

const (
	Empty   SchemaType = "" // OneOf, AnyOf, AllOf.
	Object  SchemaType = "object"
	Array   SchemaType = "array"
	Integer SchemaType = "integer"
	Number  SchemaType = "number"
	String  SchemaType = "string"
	Boolean SchemaType = "boolean"
)

// Schema is an OpenAPI JSON Schema.
type Schema struct {
	Type        SchemaType
	Format      Format // Schema format, optional.
	Description string // Schema description, optional.
	Ref         string // Whether schema is referenced.

	Item       *Schema       // Only for Array.
	Enum       []interface{} // Only for Enum.
	Properties []Property    // Only for Object.

	Nullable bool // Whether schema is nullable or not. Any types.

	OneOf []*Schema
	AnyOf []*Schema
	AllOf []*Schema

	// Numeric validation (Integer, Number).
	Maximum          *int64
	ExclusiveMaximum bool
	Minimum          *int64
	ExclusiveMinimum bool
	MultipleOf       *int

	// String validation.
	MaxLength *uint64
	MinLength *int64
	Pattern   string

	// Array validation.
	MaxItems    *uint64
	MinItems    *uint64
	UniqueItems bool

	// Struct validation.
	MaxProperties *uint64
	MinProperties *uint64
}

// Property is an OpenAPI JSON Schema Object property.
type Property struct {
	Name     string  // Field name.
	Schema   *Schema // Field schema.
	Optional bool    // Whether the field is required or not.
}
