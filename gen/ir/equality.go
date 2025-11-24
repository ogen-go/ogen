package ir

// EqualityMethodSpec describes how to generate an Equal method for a type.
type EqualityMethodSpec struct {
	// TypeName is the Go type name (e.g., "WorkflowStatus")
	TypeName string

	// Fields describes each field's equality comparison logic
	Fields []FieldEqualitySpec

	// NeedsDepthTracking indicates if this type contains nested objects
	// requiring depth limit enforcement
	NeedsDepthTracking bool

	// MaxDepth is the configured maximum nesting depth (default: 10)
	MaxDepth int
}

// FieldEqualitySpec describes how to compare a single field.
type FieldEqualitySpec struct {
	// FieldName is the struct field name (e.g., "ID", "Description")
	FieldName string

	// FieldType categorizes the field for comparison logic selection
	FieldType FieldTypeCategory

	// GoType is the full Go type (e.g., "string", "OptString", "*NestedObject")
	GoType string

	// IsNested indicates if this field is a nested object requiring recursive Equal() call
	IsNested bool

	// IsMap indicates if the field (or wrapped value for Optional/Nullable) is a map type
	IsMap bool

	// IsArray indicates if the field (or wrapped value for Optional/Nullable) is an array type
	IsArray bool

	// IsArrayOfStructs indicates if this is an array field where elements are structs
	IsArrayOfStructs bool

	// IsArrayOfNullable indicates if this is an array field where elements are nullable wrappers (NilT)
	IsArrayOfNullable bool

	// IsByteSlice indicates if this field is a byte slice ([]byte / jx.Raw)
	IsByteSlice bool
}

// FieldTypeCategory classifies fields for equality comparison.
type FieldTypeCategory int

const (
	FieldTypePrimitive    FieldTypeCategory = iota // string, int, bool, etc.
	FieldTypeOptional                              // OptT, OptNilT
	FieldTypeNullable                              // NilT
	FieldTypePointer                               // *T
	FieldTypeNestedObject                          // struct with generated Equal()
	FieldTypeArray                                 // []T
	FieldTypeMap                                   // map[K]V
)

// HashMethodSpec describes how to generate a Hash method for a type.
type HashMethodSpec struct {
	// TypeName is the Go type name
	TypeName string

	// Fields describes each field's hashing logic
	Fields []FieldHashSpec

	// UsesNestedHash indicates if this type contains nested objects
	// that provide their own Hash() methods
	UsesNestedHash bool
}

// FieldHashSpec describes how to hash a single field.
type FieldHashSpec struct {
	// FieldName is the struct field name
	FieldName string

	// FieldType categorizes the field for hashing logic selection
	FieldType FieldTypeCategory

	// GoType is the full Go type
	GoType string

	// IsNested indicates if this field has a Hash() method
	IsNested bool
}
