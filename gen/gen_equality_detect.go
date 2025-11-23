package gen

import "github.com/ogen-go/ogen/gen/ir"

const (
	prefixOpt  = "Opt"
	prefixNil  = "Nil"
	fieldValue = "Value"
)

// collectEqualitySpecs identifies types that require Equal() and Hash() methods
// for complex uniqueItems validation.
func (g *Generator) collectEqualitySpecs() {
	// Iterate through all types to find arrays with complex uniqueItems
	visited := make(map[string]bool)
	for _, t := range g.tstorage.types {
		g.collectFromType(t, visited)
	}
}

// collectFromType recursively checks a type and its fields for uniqueItems arrays
func (g *Generator) collectFromType(t *ir.Type, visited map[string]bool) {
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
			}
			// Collect specs for all nested types within this item
			g.collectNestedTypes(t.Item, visited)
		}
	}

	// Recursively check fields of structs
	if t.Kind == ir.KindStruct {
		for _, field := range t.Fields {
			g.collectFromType(field.Type, visited)
		}
	}

	// Check sum type variants
	if t.Kind == ir.KindSum {
		for _, variant := range t.SumOf {
			g.collectFromType(variant, visited)
		}
	}
}

// collectNestedTypes recursively collects all nested types that need Equal/Hash methods
func (g *Generator) collectNestedTypes(t *ir.Type, visited map[string]bool) {
	if t == nil {
		return
	}

	// Prevent infinite recursion on circular dependencies
	if t.Name != "" && visited[t.Name] {
		return
	}

	switch t.Kind {
	case ir.KindGeneric:
		// Generic types like OptT, NilT - unwrap to the underlying type
		if t.GenericOf != nil {
			g.collectNestedTypes(t.GenericOf, visited)
		}
		return

	case ir.KindStruct:
		// Check if it's an Optional/Nullable wrapper - unwrap and recurse
		name := t.Name
		if len(name) > 3 && (name[:3] == prefixOpt || name[:3] == prefixNil) {
			for _, field := range t.Fields {
				if field.Name == fieldValue {
					// Recursively process the wrapped type (don't mark wrapper as visited)
					g.collectNestedTypes(field.Type, visited)
					return
				}
			}
			return
		}

		// Mark as visited to prevent infinite recursion
		if t.Name != "" {
			visited[t.Name] = true
		}

		// Check if already collected
		alreadyExists := false
		for _, existing := range g.equalitySpecs {
			if existing.TypeName == t.Name {
				alreadyExists = true
				break
			}
		}

		// Regular struct - create spec if not already exists
		if !alreadyExists {
			spec := g.createEqualityMethodSpec(t)
			if spec != nil {
				g.equalitySpecs = append(g.equalitySpecs, spec)
				// Debug: log collection
			}
		}

		// Recurse into all fields to find more nested types
		// Debug: log field traversal
		for _, field := range t.Fields {
			g.collectNestedTypes(field.Type, visited)
		}

	case ir.KindArray:
		// Arrays might contain nested objects
		if t.Item != nil {
			g.collectNestedTypes(t.Item, visited)
		}

	case ir.KindAlias:
		if t.AliasTo != nil {
			g.collectNestedTypes(t.AliasTo, visited)
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
			// Unwrap to get the actual type for GoType
			unwrapped := unwrapOptional(field.Type)
			goType := field.Type.Go()
			isMap := false
			isArray := false
			isArrayOfStructs := false
			isByteSlice := false

			if unwrapped != nil && unwrapped != field.Type {
				goType = unwrapped.Go() // Use unwrapped type for better type detection
				isMap = unwrapped.Kind == ir.KindMap
				isArray = unwrapped.Kind == ir.KindArray
			}

			// Check if this is a byte slice (either direct or wrapped)
			if field.Type.Kind == ir.KindPrimitive && field.Type.Primitive == ir.ByteSlice {
				isByteSlice = true
			}
			if unwrapped != nil && unwrapped.Kind == ir.KindPrimitive && unwrapped.Primitive == ir.ByteSlice {
				isByteSlice = true
			}
			// Also check by GoType string for jx.Raw or []byte
			if goType == "jx.Raw" || goType == "[]byte" {
				isByteSlice = true
			}

			// Check if this is an array (either direct or wrapped in Optional/Nullable)
			if field.Type.Kind == ir.KindArray {
				isArray = true
			}

			// Check if this is an array of structs or nullable wrappers (either direct or wrapped)
			isArrayOfNullable := false
			arrayType := field.Type
			if arrayType.Kind == ir.KindArray && arrayType.Item != nil {
				// Check if item is a nullable wrapper (can be Generic or Struct)
				itemName := arrayType.Item.Name
				if len(itemName) > 3 && itemName[:3] == "Nil" &&
					(arrayType.Item.Kind == ir.KindStruct || arrayType.Item.Kind == ir.KindGeneric) {
					isArrayOfNullable = true
					// Also check if the VALUE inside the nullable wrapper is a struct
					innerType := unwrapOptional(arrayType.Item)
					if innerType != nil && innerType.Kind == ir.KindStruct {
						isArrayOfStructs = true
					}
				} else {
					// Direct struct array (not wrapped in nullable)
					itemType := unwrapOptional(arrayType.Item)
					if itemType != nil && itemType.Kind == ir.KindStruct {
						isArrayOfStructs = true
					}
				}
			} else if unwrapped != nil && unwrapped.Kind == ir.KindArray && unwrapped.Item != nil {
				// Check wrapped array (OptT[[]Struct] or NilT[[]Struct])
				itemName := unwrapped.Item.Name
				if len(itemName) > 3 && itemName[:3] == "Nil" &&
					(unwrapped.Item.Kind == ir.KindStruct || unwrapped.Item.Kind == ir.KindGeneric) {
					isArrayOfNullable = true
					// Also check if the VALUE inside the nullable wrapper is a struct
					innerType := unwrapOptional(unwrapped.Item)
					if innerType != nil && innerType.Kind == ir.KindStruct {
						isArrayOfStructs = true
					}
				} else {
					// Direct struct array (not wrapped in nullable)
					itemType := unwrapOptional(unwrapped.Item)
					if itemType != nil && itemType.Kind == ir.KindStruct {
						isArrayOfStructs = true
					}
				}
			}

			fieldSpec := ir.FieldEqualitySpec{
				FieldName:         field.Name,
				FieldType:         categorizeFieldType(field.Type),
				GoType:            goType,
				IsNested:          isNestedObject(field.Type),
				IsMap:             isMap,
				IsArray:           isArray,
				IsArrayOfStructs:  isArrayOfStructs,
				IsArrayOfNullable: isArrayOfNullable,
				IsByteSlice:       isByteSlice,
			}
			spec.Fields = append(spec.Fields, fieldSpec)
		}
	}

	return spec
}

// hasNestedObjects checks if a type contains nested objects requiring depth tracking
// For simplicity and consistency, all struct types that will have Equal() methods
// should have depth tracking. This ensures uniform Equal() signatures.
func hasNestedObjects(t *ir.Type) bool {
	if t == nil {
		return false
	}

	// All struct types get depth tracking for consistent Equal() signatures
	// This simplifies code generation and calling conventions
	return t.Kind == ir.KindStruct
}

// unwrapOptional unwraps Generic optional/nullable types to get the underlying type
func unwrapOptional(t *ir.Type) *ir.Type {
	if t == nil {
		return nil
	}

	// Unwrap generic types (OptT, NilT)
	if t.Kind == ir.KindGeneric && t.GenericOf != nil {
		return t.GenericOf
	}

	// Unwrap struct-based optional wrappers
	if t.Kind == ir.KindStruct && len(t.Name) > 3 && (t.Name[:3] == prefixOpt || t.Name[:3] == prefixNil) {
		for _, field := range t.Fields {
			if field.Name == fieldValue {
				return field.Type
			}
		}
	}

	return t
}

// isNestedObject checks if a type is a nested object (struct)
func isNestedObject(t *ir.Type) bool {
	if t == nil {
		return false
	}

	switch t.Kind {
	case ir.KindGeneric:
		// Generic types like OptT, NilT - check if they wrap a nested object
		if t.GenericOf != nil {
			return isNestedObject(t.GenericOf)
		}
		return false
	case ir.KindStruct:
		// Check if this is an Optional/Nullable wrapper
		name := t.Name
		if len(name) > 3 && (name[:3] == prefixOpt || name[:3] == prefixNil) {
			// Look for a Value field that is a nested object
			for _, field := range t.Fields {
				if field.Name == fieldValue {
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

	case ir.KindGeneric:
		// Generic types like OptT, NilT - check by name
		name := t.Name
		if len(name) > 3 && name[:3] == prefixOpt {
			return ir.FieldTypeOptional
		}
		if len(name) > 3 && name[:3] == prefixNil {
			return ir.FieldTypeNullable
		}
		// Other generic types - check the underlying type
		if t.GenericOf != nil {
			return categorizeFieldType(t.GenericOf)
		}
		return ir.FieldTypePrimitive

	case ir.KindStruct:
		// Check if this is an optional/nullable wrapper type
		name := t.Name
		if len(name) > 3 && name[:3] == prefixOpt {
			return ir.FieldTypeOptional
		}
		if len(name) > 3 && name[:3] == prefixNil {
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
