package gen

import (
	"github.com/ogen-go/ogen/gen/ir"
)

// collectEqualitySpecs identifies types that require Equal() and Hash() methods
// for complex uniqueItems validation.
func (g *Generator) collectEqualitySpecs() error {
	// Iterate through all types to find arrays with complex uniqueItems
	for _, t := range g.tstorage.types {
		g.collectFromType(t)
	}
	return nil
}

// collectFromType recursively checks a type and its fields for uniqueItems arrays
func (g *Generator) collectFromType(t *ir.Type) {
	if t == nil {
		return
	}

	// Check if this is an array with unique items
	if t.Kind == ir.KindArray && t.Schema != nil && t.Schema.UniqueItems {
		if needsEqualityMethod(t.Item) {
			// Create spec for the item type
			spec := g.createEqualityMethodSpec(t.Item)
			if spec != nil {
				g.equalitySpecs = append(g.equalitySpecs, spec)
				// Also collect specs for all nested types within this item
				g.collectNestedTypes(t.Item)
			}
		}
	}

	// Recursively check fields of structs
	if t.Kind == ir.KindStruct {
		for _, field := range t.Fields {
			g.collectFromType(field.Type)
		}
	}

	// Check sum type variants
	if t.Kind == ir.KindSum {
		for _, variant := range t.SumOf {
			g.collectFromType(variant)
		}
	}
}

// collectNestedTypes recursively collects all nested types that need Equal/Hash methods
func (g *Generator) collectNestedTypes(t *ir.Type) {
	if t == nil {
		return
	}

	// Skip if already collected
	for _, existing := range g.equalitySpecs {
		if existing.TypeName == t.Name {
			return
		}
	}

	switch t.Kind {
	case ir.KindStruct:
		// Check if it's an Optional/Nullable wrapper - look inside
		name := t.Name
		if len(name) > 3 && (name[:3] == "Opt" || name[:3] == "Nil") {
			for _, field := range t.Fields {
				if field.Name == "Value" {
					g.collectNestedTypes(field.Type)
					return
				}
			}
		}

		// Regular struct - create spec and recurse into fields
		spec := g.createEqualityMethodSpec(t)
		if spec != nil {
			g.equalitySpecs = append(g.equalitySpecs, spec)
		}

		// Recurse into all fields
		for _, field := range t.Fields {
			if isNestedObject(field.Type) {
				g.collectNestedTypes(field.Type)
			}
		}

	case ir.KindAlias:
		if t.AliasTo != nil {
			g.collectNestedTypes(t.AliasTo)
		}
	}
}

// needsEqualityMethod determines if a type requires generated Equal() method
func needsEqualityMethod(t *ir.Type) bool {
	if t == nil {
		return false
	}

	switch t.Kind {
	case ir.KindStruct:
		// Structs always need custom equality
		return true
	case ir.KindAlias:
		// Check if the underlying type needs equality
		if t.AliasTo != nil {
			return needsEqualityMethod(t.AliasTo)
		}
		return false
	case ir.KindArray, ir.KindMap:
		// Arrays and maps of complex types need equality
		return true
	default:
		// Primitives can use == operator
		return false
	}
}

// createEqualityMethodSpec creates an EqualityMethodSpec for a given type
func (g *Generator) createEqualityMethodSpec(t *ir.Type) *ir.EqualityMethodSpec {
	if t == nil || t.Name == "" {
		return nil
	}

	// Check if we've already created a spec for this type
	for _, existing := range g.equalitySpecs {
		if existing.TypeName == t.Name {
			return nil // Already tracked
		}
	}

	spec := &ir.EqualityMethodSpec{
		TypeName:           t.Name,
		Fields:             []ir.FieldEqualitySpec{},
		NeedsDepthTracking: hasNestedObjects(t),
		MaxDepth:           10, // Default from clarifications
	}

	// Populate fields for struct types
	if t.Kind == ir.KindStruct {
		for _, field := range t.Fields {
			fieldSpec := ir.FieldEqualitySpec{
				FieldName: field.Name,
				FieldType: categorizeFieldType(field.Type),
				GoType:    field.Type.Go(),
				IsNested:  isNestedObject(field.Type),
			}
			spec.Fields = append(spec.Fields, fieldSpec)
		}
	}

	return spec
}

// hasNestedObjects checks if a type contains nested objects requiring depth tracking
func hasNestedObjects(t *ir.Type) bool {
	if t == nil {
		return false
	}

	if t.Kind != ir.KindStruct {
		return false
	}

	for _, field := range t.Fields {
		if isNestedObject(field.Type) {
			return true
		}
		// Check for arrays or maps of nested objects
		if field.Type.Kind == ir.KindArray && isNestedObject(field.Type.Item) {
			return true
		}
	}

	return false
}

// isNestedObject checks if a type is a nested object (struct)
func isNestedObject(t *ir.Type) bool {
	if t == nil {
		return false
	}

	switch t.Kind {
	case ir.KindStruct:
		// Check if this is an Optional/Nullable wrapper
		name := t.Name
		if len(name) > 3 && (name[:3] == "Opt" || name[:3] == "Nil") {
			// Look for a Value field that is a nested object
			for _, field := range t.Fields {
				if field.Name == "Value" {
					return isNestedObject(field.Type)
				}
			}
			return false
		}
		// Regular struct - it's a nested object
		return true
	case ir.KindAlias:
		if t.AliasTo != nil {
			return isNestedObject(t.AliasTo)
		}
		return false
	default:
		return false
	}
}

// categorizeFieldType maps an IR type to a FieldTypeCategory
func categorizeFieldType(t *ir.Type) ir.FieldTypeCategory {
	if t == nil {
		return ir.FieldTypePrimitive
	}

	switch t.Kind {
	case ir.KindPrimitive:
		// Check for optional/nullable wrappers
		switch t.Primitive {
		case ir.String, ir.Int, ir.Int8, ir.Int16, ir.Int32, ir.Int64,
			ir.Uint, ir.Uint8, ir.Uint16, ir.Uint32, ir.Uint64,
			ir.Float32, ir.Float64, ir.Bool, ir.ByteSlice, ir.Duration, ir.Time, ir.UUID, ir.URL:
			return ir.FieldTypePrimitive
		default:
			return ir.FieldTypePrimitive
		}

	case ir.KindStruct:
		// Check if this is an optional/nullable wrapper type
		name := t.Name
		if len(name) > 3 && name[:3] == "Opt" {
			return ir.FieldTypeOptional
		}
		if len(name) > 3 && name[:3] == "Nil" {
			return ir.FieldTypeNullable
		}
		// Regular nested object
		return ir.FieldTypeNestedObject

	case ir.KindArray:
		return ir.FieldTypeArray

	case ir.KindMap:
		return ir.FieldTypeMap

	case ir.KindPointer:
		return ir.FieldTypePointer

	case ir.KindAlias:
		// For aliases, check the underlying type
		if t.AliasTo != nil {
			return categorizeFieldType(t.AliasTo)
		}
		return ir.FieldTypePrimitive

	default:
		return ir.FieldTypePrimitive
	}
}
