package ast

type SchemaType string

const (
	Object  SchemaType = "object"
	Array   SchemaType = "array"
	Integer SchemaType = "integer"
	Number  SchemaType = "number"
	String  SchemaType = "string"
	Boolean SchemaType = "boolean"
)

type Schema struct {
	Type        SchemaType
	Format      string
	Description string
	Ref         string

	Item       *Schema
	EnumValues []interface{}
	Fields     []SchemaField

	Nullable bool

	// Numeric
	MultipleOf       *int
	Maximum          *int64
	ExclusiveMaximum bool
	Minimum          *int64
	ExclusiveMinimum bool

	// String
	MaxLength *uint64
	MinLength *int64
	Pattern   string

	// Array
	MaxItems    *uint64
	MinItems    *uint64
	UniqueItems bool

	// Struct
	MaxProperties *uint64
	MinProperties *uint64
}

type SchemaField struct {
	Name     string
	Schema   *Schema
	Optional bool
}
