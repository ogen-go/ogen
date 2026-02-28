package gen

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"slices"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

// fieldSignature represents a field's discriminating characteristics (name + type).
//
// This enables type-based field discrimination: fields with the same name but different
// types are not considered "common" and can be used for discrimination.
//
// For example, if VariantA has {id: string} and VariantB has {id: integer}, the "id"
// field can help discriminate between them even though the field names are identical.
type fieldSignature struct {
	name   string
	typeID string
}

const (
	typeIDAny     = "any"
	typeIDBoolean = "boolean"
	typeIDInteger = "integer"
	typeIDNumber  = "number"
	typeIDString  = "string"
	typeIDNull    = "null"
	typeIDObject  = "object"
	typeIDSum     = "sum"
	typeIDAlias   = "alias"
	typeIDPointer = "pointer"

	// jxTypeArray is the string representation of jx.Array for template generation.
	jxTypeArray = "jx.Array"
)

// jxTypeForFieldType returns the jx.Type constant name for runtime type checking.
// Returns empty string if the type is not distinguishable at JSON level.
func jxTypeForFieldType(ft *ir.Type) string {
	typeID := getFieldTypeID(ft)
	const strType = "jx.String"
	switch {
	case typeID == typeIDBoolean:
		return "jx.Bool"
	case typeID == typeIDInteger, typeID == typeIDNumber:
		if s := ft.Schema; s != nil && s.Type == jsonschema.String {
			// TODO(tdakkota): properly figure out JSON type.
			return strType
		}
		return "jx.Number"
	case typeID == typeIDString:
		return strType
	case typeID == typeIDNull:
		return "jx.Null"
	case typeID == typeIDObject:
		return "jx.Object"
	case strings.HasPrefix(typeID, "array["):
		return jxTypeArray
	case strings.HasPrefix(typeID, "map["):
		return "jx.Object"
	case strings.HasPrefix(typeID, "enum_"):
		// Enums serialize as strings in JSON
		return strType
	default:
		return ""
	}
}

// getArrayElementTypeInfo extracts element type information from an array type ID.
// Returns the element type ID and its corresponding jx.Type.
// For non-array types, returns empty strings.
func getArrayElementTypeInfo(t *ir.Type) (elementTypeID, elementJxType string) {
	typeID := getFieldTypeID(t)
	if !strings.HasPrefix(typeID, "array[") || t.Item == nil {
		return "", ""
	}

	// Extract element type: "array[string]" -> "string"
	elementTypeID = strings.TrimPrefix(typeID, "array[")
	elementTypeID = strings.TrimSuffix(elementTypeID, "]")

	// Get the jx.Type for the element
	elementJxType = jxTypeForFieldType(t.Item)

	return elementTypeID, elementJxType
}

// getFieldTypeID returns a type identifier for discrimination purposes.
// Fields with the same name but different typeIDs can discriminate variants.
func getFieldTypeID(t *ir.Type) string {
	if t == nil {
		return typeIDAny
	}

	// Unwrap optionals and nullables to get the base type
	base := t
	for base.IsGeneric() {
		if v := base.GenericOf; v != nil {
			base = v
			continue
		}
		break
	}

	switch base.Kind {
	case ir.KindAny:
		return typeIDAny
	case ir.KindPrimitive:
		switch base.Primitive {
		case ir.Bool:
			return typeIDBoolean
		case ir.Int, ir.Int8, ir.Int16, ir.Int32, ir.Int64,
			ir.Uint, ir.Uint8, ir.Uint16, ir.Uint32, ir.Uint64:
			return typeIDInteger
		case ir.Float32, ir.Float64:
			return typeIDNumber
		case ir.String, ir.ByteSlice:
			return typeIDString
		case ir.Null:
			return typeIDNull
		default:
			return fmt.Sprintf("primitive_%s", base.Primitive)
		}
	case ir.KindArray:
		itemID := typeIDAny
		if base.Item != nil {
			itemID = getFieldTypeID(base.Item)
		}
		return fmt.Sprintf("array[%s]", itemID)
	case ir.KindEnum:
		// Enums are distinct from their underlying types
		return fmt.Sprintf("enum_%s", base.Name)
	case ir.KindStruct:
		return typeIDObject
	case ir.KindMap:
		itemID := typeIDAny
		if base.Item != nil {
			itemID = getFieldTypeID(base.Item)
		}
		return fmt.Sprintf("map[%s]", itemID)
	case ir.KindSum:
		return typeIDSum
	case ir.KindAlias:
		if base.AliasTo != nil {
			return getFieldTypeID(base.AliasTo)
		}
		return typeIDAlias
	case ir.KindPointer:
		if base.PointerTo != nil {
			return getFieldTypeID(base.PointerTo)
		}
		return typeIDPointer
	default:
		return string(base.Kind)
	}
}

func canUseTypeDiscriminator(sum []*ir.Type, isOneOf bool) bool {
	var (
		// Collect map of variant kinds.
		typeMap      = map[string]struct{}{}
		collectTypes func(sum []*ir.Type)

		getType = func(t *ir.Type) string {
			typ := t.JSON().Type()
			if s := t.Schema; s != nil && s.Type == jsonschema.Integer && typ == "Number" {
				// Special case for anyOf with integer and number.
				if !isOneOf {
					typ = "Integer"
				}
			}
			return typ
		}
	)
	collectTypes = func(sum []*ir.Type) {
		for _, variant := range sum {
			typ := getType(variant)
			if typ == "" {
				if variant.IsSum() {
					collectTypes(variant.SumOf)
				}
				continue
			}
			typeMap[typ] = struct{}{}
		}
	}

	var hasSumVariant bool
	for _, s := range sum {
		typ := getType(s)
		switch {
		case s.IsSum():
			hasSumVariant = true
			switch s.JSON().Sum().Type {
			case ir.SumJSONDiscriminator, ir.SumJSONFields:
				typeMap["Object"] = struct{}{}
			case ir.SumJSONPrimitive, ir.SumJSONTypeDiscriminator:
				collectTypes(s.SumOf)
			}
			continue
		case typ == "":
			// Cannot make type discriminator with Any.
			return false
		}

		if _, ok := typeMap[typ]; ok {
			// Type kind is not unique, so we cannot distinguish variants by type.
			return false
		}
		typeMap[typ] = struct{}{}
	}

	_, hasInteger := typeMap["Integer"]
	_, hasNumber := typeMap["Number"]
	if hasInteger && hasNumber && hasSumVariant {
		// TODO(tdakkota): Do not allow type discriminator for nested sum types with integer and
		// 	number variants at the same time. We can add support for this later, but it's not trivial.
		return false
	}
	return true
}

func ensureNoInfiniteRecursion(parent *jsonschema.Schema) error {
	var do func(map[jsonschema.Ref]struct{}, []*jsonschema.Schema) error
	do = func(ctx map[jsonschema.Ref]struct{}, schemas []*jsonschema.Schema) error {
		for i, s := range schemas {
			if s == nil {
				// Just skip nil schemas. We handle them later.
				continue
			}
			if ref := s.Ref; !ref.IsZero() {
				if _, ok := ctx[ref]; ok {
					err := errors.Errorf("reference %q [%d] leads to infinite recursion", ref, i)

					pos, ok := s.Position()
					if !ok {
						return err
					}
					return &location.Error{
						File: s.File(),
						Pos:  pos,
						Err:  err,
					}
				}
				ctx[ref] = struct{}{}
			}
			switch {
			case len(s.OneOf) > 0:
				if err := do(ctx, s.OneOf); err != nil {
					return err
				}
			case len(s.AllOf) > 0:
				if err := do(ctx, s.AllOf); err != nil {
					return err
				}
			case len(s.AnyOf) > 0:
				if err := do(ctx, s.AnyOf); err != nil {
					return err
				}
			}
			delete(ctx, s.Ref)
		}
		return nil
	}

	return do(map[jsonschema.Ref]struct{}{}, []*jsonschema.Schema{parent})
}

func (g *schemaGen) collectSumVariants(
	name string,
	schemas []*jsonschema.Schema,
) (sum []*ir.Type, _ error) {
	// TODO(tdakkota): convert oneOf+null into generic

	for _, s := range schemas {
		if s != nil && s.Nullable {
			nullT := ir.Primitive(ir.Null, nil)
			nullT.Name = "Null"
			sum = append(sum, nullT)
			break
		}
	}

	names := make(map[string]struct{}, len(schemas))
	for i, s := range schemas {
		// generate without boxing because:
		// 1) sum variant cannot be optional
		// 2) if sum variant is nullable - null type already added into sum
		t, err := g.generate2(fmt.Sprintf("%s%d", name, i), s)
		if err != nil {
			return nil, errors.Wrapf(err, "oneOf[%d]", i)
		}

		t.Name = variantFieldName(t)
		if _, ok := names[t.Name]; ok {
			return nil, errors.Wrap(&ErrNotImplemented{
				Name: "sum types with same names",
			}, name)
		}

		names[t.Name] = struct{}{}
		sum = append(sum, t)
	}
	return sum, nil
}

func schemaName(k jsonschema.Ref) (string, bool) {
	_, after, ok := strings.Cut(k.Ptr, "#/")
	if !ok || after == "" {
		return "", false
	}
	return path.Base(after), true
}

// handleExplicitDiscriminator processes explicit discriminator mappings for both oneOf and anyOf.
// Returns true if discriminator was handled, false if no discriminator present.
func (g *schemaGen) handleExplicitDiscriminator(sum *ir.Type, schema *jsonschema.Schema, variants []*jsonschema.Schema) (bool, error) {
	d := schema.Discriminator
	if d == nil {
		return false, nil
	}

	propName := d.PropertyName

	// Build mappings and collect keys
	var mappingKeys []string
	for k, v := range d.Mapping {
		var found bool
		for i, s := range sum.SumOf {
			if !s.Is(ir.KindStruct, ir.KindMap) {
				return false, errors.Wrapf(&ErrNotImplemented{"unsupported sum type variant"}, "%q", s.Kind)
			}
			vschema := s.Schema
			if vschema == nil {
				vschema = variants[i]
			}

			var discriminatorType *ir.Type
			for _, field := range s.Fields {
				if field.Spec != nil && field.Spec.Name == propName {
					discriminatorType = field.Type
				}
			}

			if vschema.Ref == v.Ref {
				found = true
				sum.SumSpec.Mapping = append(sum.SumSpec.Mapping, ir.SumSpecMap{
					Key:               k,
					Type:              s,
					DiscriminatorType: discriminatorType,
				})
				mappingKeys = append(mappingKeys, k)
				break
			}
		}
		if !found {
			return false, errors.Errorf("mapping %q: variant %q not found", k, v.Ref)
		}
	}

	// Only proceed if we have explicit mappings
	if len(mappingKeys) == 0 {
		return false, nil
	}

	// Validate: Check if discriminator field uses value-based discrimination
	// Collect the discriminator field type from each variant
	discriminatorFieldTypes := make(map[string]string) // variant name -> jxType
	for _, mapping := range sum.SumSpec.Mapping {
		variant := mapping.Type
		// Find the discriminator field in this variant
		for _, f := range variant.JSON().Fields() {
			if f.Tag.JSON == propName {
				jxType := jxTypeForFieldType(f.Type)
				discriminatorFieldTypes[variant.Name] = jxType
				break
			}
		}
	}

	// Check if all discriminator fields have the same empty jxType (value-based discrimination)
	if len(discriminatorFieldTypes) > 1 {
		firstJxType := ""
		allSameAndEmpty := true
		for _, jxType := range discriminatorFieldTypes {
			if firstJxType == "" {
				firstJxType = jxType
			} else if jxType != firstJxType {
				allSameAndEmpty = false
				break
			}
			if jxType != "" {
				allSameAndEmpty = false
			}
		}

		if allSameAndEmpty {
			variantNames := make([]string, 0, len(discriminatorFieldTypes))
			for name := range discriminatorFieldTypes {
				variantNames = append(variantNames, name)
			}
			return true, errors.Wrapf(
				&ErrNotImplemented{Name: "value-based discriminator"},
				"discriminator field %q: all variants have same JSON type (cannot discriminate by type), variants: %v",
				propName, variantNames,
			)
		}
	}

	// Set discriminator only if we have mappings
	sum.SumSpec.Discriminator = propName

	// Generate names using the helper
	nameGen, err := discriminatorMappingNameGen(sum.Name, mappingKeys)
	if err != nil {
		return true, err
	}

	// Generate Go variable name for each mapping Key
	for idx, m := range sum.SumSpec.Mapping {
		name, err := nameGen(m.Key, idx)
		if err != nil {
			return true, errors.Wrapf(err, "variant %q", m.Key)
		}
		sum.SumSpec.Mapping[idx].Name = name
	}

	slices.SortStableFunc(sum.SumSpec.Mapping, func(a, b ir.SumSpecMap) int {
		return strings.Compare(a.Key, b.Key)
	})
	return true, nil
}

func (g *schemaGen) anyOf(name string, schema *jsonschema.Schema, side bool) (*ir.Type, error) {
	if err := ensureNoInfiniteRecursion(schema); err != nil {
		return nil, err
	}

	var regSchema *jsonschema.Schema
	if !side {
		regSchema = schema
	}
	sum := g.regtype(name, &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: regSchema,
	})
	{
		variants, err := g.collectSumVariants(name, schema.AnyOf)
		if err != nil {
			return nil, errors.Wrap(err, "collect variants")
		}
		sum.SumOf = variants
	}

	// Here we try to create sum type from anyOf for variants with JSON type-based discriminator.
	if canUseTypeDiscriminator(sum.SumOf, false) {
		sum.SumSpec.TypeDiscriminator = true

		for _, v := range sum.SumOf {
			switch v.Kind {
			case ir.KindPrimitive, ir.KindEnum:
				switch {
				case v.IsInteger():
					if !v.Validators.Int.Set() {
						if err := v.Validators.SetInt(schema); err != nil {
							return nil, errors.Wrap(err, "int validator")
						}
					}
				case v.IsFloat():
					if !v.Validators.Float.Set() {
						if err := v.Validators.SetFloat(schema); err != nil {
							return nil, errors.Wrap(err, "float validator")
						}
					}
				case v.IsDecimal():
					if !v.Validators.Decimal.Set() {
						if err := v.Validators.SetDecimal(schema); err != nil {
							return nil, errors.Wrap(err, "decimal validator")
						}
					}
				case !v.Validators.String.Set():
					if err := v.Validators.SetString(schema); err != nil {
						return nil, errors.Wrap(err, "string validator")
					}
				}
			case ir.KindArray:
				if !v.Validators.Array.Set() {
					v.Validators.SetArray(schema)
				}
			case ir.KindMap, ir.KindStruct:
				if !v.Validators.Object.Set() {
					v.Validators.SetObject(schema)
				}
			}
		}
		return sum, nil
	}

	// Check for explicit discriminator
	if handled, err := g.handleExplicitDiscriminator(sum, schema, schema.AnyOf); handled {
		return sum, err
	}

	return nil, &ErrNotImplemented{"complex anyOf"}
}

func (g *schemaGen) oneOf(name string, schema *jsonschema.Schema, side bool) (*ir.Type, error) {
	if err := ensureNoInfiniteRecursion(schema); err != nil {
		return nil, err
	}

	// Collect variants first to enable early validation
	variants, err := g.collectSumVariants(name, schema.OneOf)
	if err != nil {
		return nil, errors.Wrap(err, "collect variants")
	}

	// Early validation for explicit discriminator (before regtype)
	// This prevents broken types from being registered when discriminator is invalid
	if schema.Discriminator != nil {
		// Quick validation: check if this would be value-based discrimination
		d := schema.Discriminator
		propName := d.PropertyName

		// Only validate if there are explicit mappings
		if len(d.Mapping) > 0 {
			discriminatorFieldTypes := make(map[string]string)
			for _, variant := range variants {
				for _, f := range variant.JSON().Fields() {
					if f.Tag.JSON == propName {
						jxType := jxTypeForFieldType(f.Type)
						discriminatorFieldTypes[variant.Name] = jxType
						break
					}
				}
			}

			// Check if all discriminator fields have same empty jxType (value-based)
			if len(discriminatorFieldTypes) > 1 {
				firstJxType := ""
				allSameAndEmpty := true
				for _, jxType := range discriminatorFieldTypes {
					if firstJxType == "" {
						firstJxType = jxType
					} else if jxType != firstJxType {
						allSameAndEmpty = false
						break
					}
					if jxType != "" {
						allSameAndEmpty = false
					}
				}

				if allSameAndEmpty {
					variantNames := make([]string, 0, len(discriminatorFieldTypes))
					for name := range discriminatorFieldTypes {
						variantNames = append(variantNames, name)
					}
					return nil, errors.Wrapf(
						&ErrNotImplemented{Name: "value-based discriminator"},
						"discriminator field %q: all variants have same JSON type (cannot discriminate by type), variants: %v",
						propName, variantNames,
					)
				}
			}
		}
	}

	var regSchema *jsonschema.Schema
	if !side {
		regSchema = schema
	}
	sum := g.regtype(name, &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: regSchema,
	})
	sum.SumOf = variants

	// 1st case: explicit discriminator.
	if handled, err := g.handleExplicitDiscriminator(sum, schema, schema.OneOf); handled {
		return sum, err
	}

	// 2nd case: implicit mapping based on schema references (only if discriminator present but no explicit mappings).
	if schema.Discriminator != nil && len(sum.SumSpec.Mapping) == 0 {
		// Implicit mapping, defaults to type name.
		keys := map[string]struct{}{}
		var mappingKeys []string
		for i, s := range sum.SumOf {
			var ref jsonschema.Ref
			if s.Schema != nil {
				ref = s.Schema.Ref
			} else {
				ref = schema.OneOf[i].Ref
			}

			key, err := func() (string, error) {
				// Spec says (https://spec.openapis.org/oas/v3.1.0#discriminator-object):
				//
				// 	The expectation now is that a property with name petType MUST be present in the response payload,
				// 	and the value will correspond to the name of a schema defined in the OAS document
				//
				// What is name of a schema? Is it the last part of the pointer?
				// What if pointer part of reference is empty, like `User.json#`?
				//
				// As always, OpenAPI is not clear enough.
				key, ok := schemaName(ref)
				if !ok {
					return "", errors.Wrapf(
						&ErrNotImplemented{"complicated reference"},
						"unable to extract schema name from %s", ref,
					)
				}

				if _, ok := keys[key]; ok {
					return "", errors.Wrapf(
						&ErrNotImplemented{"duplicate mapping key"},
						"key %q", key,
					)
				}
				keys[key] = struct{}{}
				return key, nil
			}()
			if err != nil {
				return nil, errors.Wrapf(err, "mapping %q", ref)
			}

			sum.SumSpec.Mapping = append(sum.SumSpec.Mapping, ir.SumSpecMap{
				Key:  key,
				Type: s,
				// Name will be set below
			})
			mappingKeys = append(mappingKeys, key)
		}

		// Generate names using the helper (only if we have mappings)
		if len(mappingKeys) > 0 {
			// Set discriminator only if we have mappings
			sum.SumSpec.Discriminator = schema.Discriminator.PropertyName

			nameGen, err := discriminatorMappingNameGen(sum.Name, mappingKeys)
			if err != nil {
				return nil, err
			}

			// Apply generated names to mappings
			for idx, m := range sum.SumSpec.Mapping {
				name, err := nameGen(m.Key, idx)
				if err != nil {
					return nil, errors.Wrapf(err, "variant %q", m.Key)
				}
				sum.SumSpec.Mapping[idx].Name = name
			}
		}
		slices.SortStableFunc(sum.SumSpec.Mapping, func(a, b ir.SumSpecMap) int {
			return strings.Compare(a.Key, b.Key)
		})
		return sum, nil
	}

	// 3rd case: distinguish by serialization type.
	if canUseTypeDiscriminator(sum.SumOf, true) {
		sum.SumSpec.TypeDiscriminator = true
		return sum, nil
	}

	// 4th case: distinguish by unique fields (considering field types).

	// Determine unique fields for each SumOf variant.
	// We now track field signatures (name + type) instead of just names.
	uniq := map[string]map[fieldSignature]struct{}{}
	fieldNameToSignatures := map[string]map[string]map[fieldSignature]struct{}{} // variant -> fieldName -> signatures

	for _, s := range sum.SumOf {
		uniq[s.Name] = map[fieldSignature]struct{}{}
		fieldNameToSignatures[s.Name] = map[string]map[fieldSignature]struct{}{}

		if !s.Is(ir.KindStruct) {
			return nil, errors.Wrapf(&ErrNotImplemented{Name: "discriminator inference"},
				"oneOf %s: variant %s: no unique fields, "+
					"unable to parse without discriminator", sum.Name, s.Name,
			)
		}
		for _, f := range s.JSON().Fields() {
			sig := fieldSignature{
				name:   f.Tag.JSON,
				typeID: getFieldTypeID(f.Type),
			}
			uniq[s.Name][sig] = struct{}{}

			if fieldNameToSignatures[s.Name][f.Tag.JSON] == nil {
				fieldNameToSignatures[s.Name][f.Tag.JSON] = map[fieldSignature]struct{}{}
			}
			fieldNameToSignatures[s.Name][f.Tag.JSON][sig] = struct{}{}
		}
	}
	{
		// Collect field signatures that are common to at least 2 variants.
		// A field signature is common if it appears in multiple variants with the exact same name AND type.
		commonSigs := map[fieldSignature]struct{}{}
		for _, variant := range sum.SumOf {
			k := variant.Name
			sigs := uniq[k]
			for _, otherVariant := range sum.SumOf {
				otherK := otherVariant.Name
				if otherK == k {
					continue
				}
				otherSigs := uniq[otherK]
				for otherSig := range otherSigs {
					if _, has := sigs[otherSig]; has {
						// variant and otherVariant have common field signature.
						// This means same field name with same type.
						commonSigs[otherSig] = struct{}{}
					}
				}
			}
		}

		// Delete common field signatures from unique sets.
		for sig := range commonSigs {
			for _, variant := range sum.SumOf {
				delete(uniq[variant.Name], sig)
			}
		}

		// Check that at most one type has no unique fields.
		noUniqueFields := map[string]struct{}{}
		for _, variant := range sum.SumOf {
			k := variant.Name
			if len(uniq[k]) == 0 {
				// Set mapping without unique fields as default
				if len(noUniqueFields) < 1 {
					sum.SumSpec.DefaultMapping = k
				}
				noUniqueFields[k] = struct{}{}
			}
		}

		if len(noUniqueFields) > 1 {
			// Unable to deterministically select sub-schema only on fields.

			// Collect field -> variant mapping to compute fields used by multiple variants.
			fieldToVariants := map[string]map[*ir.Type]struct{}{}
			for _, variant := range sum.SumOf {
				for _, f := range variant.JSON().Fields() {
					m, ok := fieldToVariants[f.Tag.JSON]
					if !ok {
						m = map[*ir.Type]struct{}{}
						fieldToVariants[f.Tag.JSON] = m
					}
					m[variant] = struct{}{}
				}
			}

			// Collect the problematic variants and fields.
			badVariants := make([]BadVariant, 0, len(noUniqueFields))
			for _, variant := range sum.SumOf {
				if _, ok := noUniqueFields[variant.Name]; !ok {
					continue
				}

				fields := map[string][]*ir.Type{}
				for _, f := range variant.JSON().Fields() {
					for typ := range fieldToVariants[f.Tag.JSON] {
						if typ == variant {
							continue
						}
						fields[f.Tag.JSON] = append(fields[f.Tag.JSON], typ)
					}
				}
				badVariants = append(badVariants, BadVariant{
					Type:   variant,
					Fields: fields,
				})
			}

			return nil, &ErrFieldsDiscriminatorInference{
				Sum:   sum,
				Types: badVariants,
			}
		}
	}

	type sumVariant struct {
		Name   string
		Unique []string
	}
	// Build sorted list of variants with their unique field names.
	// Use alphabetical sorting to match upstream behavior and minimize diffs.
	var sortedVariants []sumVariant
	for _, s := range sum.SumOf {
		k := s.Name
		sigs := uniq[k]

		// Extract unique field names from signatures
		uniqueNames := make(map[string]struct{})
		for sig := range sigs {
			uniqueNames[sig.name] = struct{}{}
		}

		// Sort field names alphabetically
		sortedNames := make([]string, 0, len(uniqueNames))
		for name := range uniqueNames {
			sortedNames = append(sortedNames, name)
		}
		slices.Sort(sortedNames)

		sortedVariants = append(sortedVariants, sumVariant{
			Name:   k,
			Unique: sortedNames,
		})
	}
	// Sort variants alphabetically by name
	slices.SortStableFunc(sortedVariants, func(a, b sumVariant) int {
		return strings.Compare(a.Name, b.Name)
	})

	// Populate SumSpec.Unique with fields that are discriminating.
	// Also build UniqueFields map for template iteration.
	sum.SumSpec.UniqueFields = make(map[string][]ir.UniqueFieldVariant)
	for _, v := range sortedVariants {
		for _, s := range sum.SumOf {
			if s.Name != v.Name {
				continue
			}

			// Initialize UniqueFieldTypes map for runtime type checking
			if s.SumSpec.UniqueFieldTypes == nil {
				s.SumSpec.UniqueFieldTypes = make(map[string]string)
			}

			// Iterate through fields in schema order
			for _, f := range s.JSON().Fields() {
				// Check if this field name is in the unique list
				if !slices.Contains(v.Unique, f.Tag.JSON) {
					continue
				}

				// Verify the field signature is actually unique
				sig := fieldSignature{
					name:   f.Tag.JSON,
					typeID: getFieldTypeID(f.Type),
				}
				if _, ok := uniq[s.Name][sig]; !ok {
					continue
				}

				// Check if this field was already added to Unique
				alreadyAdded := false
				for _, existing := range s.SumSpec.Unique {
					if existing.Tag.JSON == f.Tag.JSON {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					s.SumSpec.Unique = append(s.SumSpec.Unique, f)
				}

				// Store expected jx.Type for runtime type checking
				jxType := jxTypeForFieldType(f.Type)
				if jxType != "" {
					s.SumSpec.UniqueFieldTypes[sig.name] = jxType
				}

				// Check if field is nullable (can be null in JSON)
				// A field is nullable if:
				// 1. It's a generic type (KindGeneric) with Nullable=true (e.g., NilString, OptNilInt)
				// 2. It's a pointer with Null semantic
				isNullable := (f.Type.IsGeneric() && f.Type.GenericVariant.Nullable) ||
					(f.Type.IsPointer() && f.Type.NilSemantic.Null())

				// Get array element type info for array element discrimination
				elemTypeID, elemJxType := getArrayElementTypeInfo(f.Type)

				// Add to UniqueFields map for template iteration
				// Include entries even when jxType is empty (simple field-name discrimination)
				sum.SumSpec.UniqueFields[f.Tag.JSON] = append(sum.SumSpec.UniqueFields[f.Tag.JSON], ir.UniqueFieldVariant{
					VariantName:        s.Name,
					VariantType:        s.Name + sum.Name,
					FieldType:          jxType,     // Empty string means no runtime type check needed
					Nullable:           isNullable, // true if field accepts null values
					ArrayElementType:   elemJxType,
					ArrayElementTypeID: elemTypeID,
				})
			}
		}
	}

	// canUseValueDiscrimination checks if a field can discriminate variants by enum values.
	// Returns true if all variants have non-overlapping enum values for this field.
	canUseValueDiscrimination := func(fieldName string, variants []ir.UniqueFieldVariant) (bool, ir.ValueDiscriminator, error) {
		if len(variants) < 2 {
			return false, ir.ValueDiscriminator{}, nil
		}

		// Collect enum values for each variant
		type variantEnumInfo struct {
			variantType string
			enumValues  map[string]bool // set of enum values
			irType      *ir.Type
		}
		variantEnums := make(map[string]*variantEnumInfo)

		for _, fv := range variants {
			// Find the IR type for this variant
			var variantIRType *ir.Type
			for _, s := range sum.SumOf {
				if s.Name+sum.Name == fv.VariantType {
					variantIRType = s
					break
				}
			}
			if variantIRType == nil {
				continue
			}

			// Find the field in this variant
			var fieldType *ir.Type
			for _, f := range variantIRType.JSON().Fields() {
				if f.Tag.JSON == fieldName {
					fieldType = f.Type
					break
				}
			}
			if fieldType == nil {
				continue
			}

			// Unwrap optionals/pointers to get base type
			baseType := fieldType
		unwrapLoop:
			for baseType.IsGeneric() || baseType.IsPointer() {
				switch {
				case baseType.IsGeneric() && baseType.GenericOf != nil:
					baseType = baseType.GenericOf
				case baseType.IsPointer() && baseType.PointerTo != nil:
					baseType = baseType.PointerTo
				default:
					break unwrapLoop
				}
			}

			// Check if it's an enum
			if baseType.Kind != ir.KindEnum {
				// Not an enum, can't use value discrimination
				return false, ir.ValueDiscriminator{}, nil
			}

			// Collect enum values
			enumValues := make(map[string]bool)
			for _, ev := range baseType.EnumVariants {
				// EnumVariant.Value is the actual value (string, int, etc.)
				// For string enums, it's already a string
				if strVal, ok := ev.Value.(string); ok {
					enumValues[strVal] = true
				} else {
					// Non-string enums (shouldn't happen for JSON strings, but be safe)
					return false, ir.ValueDiscriminator{}, nil
				}
			}

			if len(enumValues) == 0 {
				// Empty enum, can't discriminate
				return false, ir.ValueDiscriminator{}, nil
			}

			variantEnums[fv.VariantType] = &variantEnumInfo{
				variantType: fv.VariantType,
				enumValues:  enumValues,
				irType:      baseType,
			}
		}

		// Check that we found enum info for all variants
		if len(variantEnums) != len(variants) {
			// Some variants don't have enum types
			return false, ir.ValueDiscriminator{}, nil
		}

		// Check for overlapping enum values
		valueToVariants := make(map[string][]string)
		for variantType, info := range variantEnums {
			for enumValue := range info.enumValues {
				valueToVariants[enumValue] = append(valueToVariants[enumValue], variantType)
			}
		}

		// Find overlaps
		var overlappingValues []string
		for value, variantTypes := range valueToVariants {
			if len(variantTypes) > 1 {
				overlappingValues = append(overlappingValues, value)
			}
		}

		if len(overlappingValues) > 0 {
			// Build error message about overlapping values
			slices.Sort(overlappingValues)
			return false, ir.ValueDiscriminator{}, errors.Errorf(
				"field %q has overlapping enum values %v across variants; "+
					"enum values must be disjoint for automatic discrimination or use an explicit discriminator",
				fieldName,
				overlappingValues,
			)
		}

		// Build the value -> variant mapping
		valueToVariant := make(map[string]string)
		for variantType, info := range variantEnums {
			for enumValue := range info.enumValues {
				valueToVariant[enumValue] = variantType
			}
		}

		return true, ir.ValueDiscriminator{
			FieldName:      fieldName,
			ValueToVariant: valueToVariant,
		}, nil
	}

	// Validate that we can actually discriminate variants after jxType deduplication
	// This catches cases like arrays with different element types that both map to jx.Array
	//
	// We need to find at least one field that can discriminate all variants.
	// If a field has overlapping enum values, we skip it and try other fields.
	// Only fail if NO field can discriminate.

	// Track fields that need discrimination but couldn't be discriminated
	type undiscriminableField struct {
		fieldName string
		typeIDs   []string
		reason    string
	}
	var undiscriminableFields []undiscriminableField
	hasSuccessfulDiscriminator := false

	for fieldName, fieldVariants := range sum.SumSpec.UniqueFields {
		if len(fieldVariants) < 2 {
			continue // Single variant, no need to discriminate
		}

		// Count unique jxTypes for this field
		uniqueJxTypes := make(map[string]bool)
		for _, fv := range fieldVariants {
			if fv.FieldType != "" {
				uniqueJxTypes[fv.FieldType] = true
			}
		}

		// If all variants have the same jxType (or empty), try value-based or array element discrimination
		if len(uniqueJxTypes) <= 1 {
			// Try value-based discrimination (enum values)
			canUse, discriminator, err := canUseValueDiscrimination(fieldName, fieldVariants)
			if err != nil {
				// Overlapping enum values - record this but continue checking other fields
				var typeIDs []string
				for _, v := range sortedVariants {
					for _, s := range sum.SumOf {
						if s.Name != v.Name {
							continue
						}
						for _, f := range s.JSON().Fields() {
							if f.Tag.JSON == fieldName {
								typeID := getFieldTypeID(f.Type)
								typeIDs = append(typeIDs, fmt.Sprintf("%s: %s", s.Name, typeID))
								break
							}
						}
					}
				}
				undiscriminableFields = append(undiscriminableFields, undiscriminableField{
					fieldName: fieldName,
					typeIDs:   typeIDs,
					reason:    err.Error(),
				})
				continue // Try next field
			}
			if canUse {
				// Initialize map if needed
				if sum.SumSpec.ValueDiscriminators == nil {
					sum.SumSpec.ValueDiscriminators = make(map[string]ir.ValueDiscriminator)
				}
				// Store the value discriminator
				sum.SumSpec.ValueDiscriminators[fieldName] = discriminator
				// Remove from UniqueFields to avoid duplicate case statements in template
				delete(sum.SumSpec.UniqueFields, fieldName)
				hasSuccessfulDiscriminator = true
				continue // This field can discriminate, move to next field
			}

			// Value discrimination didn't work, check if array element discrimination is possible
			allArrays := true
			uniqueArrayElemTypes := make(map[string]bool)
			for _, fv := range fieldVariants {
				if fv.FieldType != jxTypeArray {
					allArrays = false
					break
				}
				if fv.ArrayElementType != "" {
					uniqueArrayElemTypes[fv.ArrayElementType] = true
				}
			}

			// If all variants are arrays with different element types, check if we have other discriminating fields
			if allArrays && len(uniqueArrayElemTypes) > 1 {
				// Array element discrimination works, but we need to check if there are other
				// unique fields to discriminate in case this array field is missing or empty.
				// If this array field is the ONLY way to discriminate, we should reject it
				// because the field might be optional and missing from the JSON.

				// Check if any variant has a unique field that exists ONLY in that variant (by name)
				hasOtherUniqueFields := false
				variantFieldNames := make(map[string]map[string]struct{}) // variant -> set of field names
				for _, variant := range sum.SumOf {
					variantFieldNames[variant.Name] = make(map[string]struct{})
					for _, f := range variant.JSON().Fields() {
						variantFieldNames[variant.Name][f.Tag.JSON] = struct{}{}
					}
				}

				// For each variant, check if it has any field name unique to it
				for _, variant := range sum.SumOf {
					for fieldName := range variantFieldNames[variant.Name] {
						isUniqueToVariant := true
						for _, otherVariant := range sum.SumOf {
							if otherVariant.Name == variant.Name {
								continue
							}
							if _, hasField := variantFieldNames[otherVariant.Name][fieldName]; hasField {
								isUniqueToVariant = false
								break
							}
						}
						if isUniqueToVariant {
							hasOtherUniqueFields = true
							break
						}
					}
					if hasOtherUniqueFields {
						break
					}
				}

				if hasOtherUniqueFields {
					continue // Can discriminate by array element type (with fallback to other fields)
				}
				// Fall through to the error below - array element discrimination alone is not sufficient
			}

			// Can't use value discrimination or array element discrimination, record for potential error
			var typeIDs []string
			for _, v := range sortedVariants {
				for _, s := range sum.SumOf {
					if s.Name != v.Name {
						continue
					}
					for _, f := range s.JSON().Fields() {
						if f.Tag.JSON == fieldName {
							typeID := getFieldTypeID(f.Type)
							typeIDs = append(typeIDs, fmt.Sprintf("%s: %s", s.Name, typeID))
							break
						}
					}
				}
			}
			undiscriminableFields = append(undiscriminableFields, undiscriminableField{
				fieldName: fieldName,
				typeIDs:   typeIDs,
				reason:    "no enum values for discrimination",
			})
		}
	}

	// If we have undiscriminable fields and no successful discriminator was found,
	// we need to fail. But if at least one field can discriminate, we're okay.
	if len(undiscriminableFields) > 0 && !hasSuccessfulDiscriminator {
		// Use the first undiscriminable field for the error message
		f := undiscriminableFields[0]
		return nil, errors.Wrapf(
			&ErrNotImplemented{Name: "type-based discrimination with same jxType"},
			"field %q cannot discriminate variants (%s): %v",
			f.fieldName,
			f.reason,
			f.typeIDs,
		)
	}

	return sum, nil
}

func (g *schemaGen) allOf(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	if err := ensureNoInfiniteRecursion(schema); err != nil {
		return nil, err
	}

	// If there is only one schema in allOf, avoid merging to keep the reference.
	if len(schema.AllOf) == 1 {
		s := schema.AllOf[0]
		if s != nil {
			return g.generate(name, s, false)
		}
	}

	mergedSchema, err := mergeNSchemes(schema.AllOf)
	if err != nil {
		return nil, err
	}

	// The reference field must not change
	mergedSchema.Ref = schema.Ref

	return g.generate(name, mergedSchema, false)
}

// shallowSchemaCopy returns a shallow copy of the given schema.
//
// If given schema is nil, nil is returned.
//
// All references in Schema are shallow copied.
func shallowSchemaCopy(s *jsonschema.Schema) *jsonschema.Schema {
	if s == nil {
		return nil
	}
	cpy := *s
	return &cpy
}

func mergeNSchemes(ss []*jsonschema.Schema) (_ *jsonschema.Schema, err error) {
	switch len(ss) {
	case 0:
		panic("unreachable")
	case 1:
		return shallowSchemaCopy(ss[0]), nil
	}

	root := ss[0]
	for i := 1; i < len(ss); i++ {
		root, err = mergeSchemes(root, ss[i])
		if err != nil {
			return nil, err
		}
	}

	return root, nil
}

func mergeSchemes(s1, s2 *jsonschema.Schema) (_ *jsonschema.Schema, err error) {
	// Helper functions for comparing validation fields.
	var (
		someU64 = func(n1, n2 *uint64, both func(n1, n2 uint64) uint64) *uint64 {
			switch {
			case n1 == nil && n2 == nil:
				return nil
			case n1 != nil && n2 == nil:
				return n1
			case n1 == nil && n2 != nil:
				return n2
			default:
				result := both(*n1, *n2)
				return &result
			}
		}

		selectMaxU64 = func(n1, n2 uint64) uint64 {
			if n1 > n2 {
				return n1
			}
			return n2
		}

		selectMinU64 = func(n1, n2 uint64) uint64 {
			if n1 < n2 {
				return n1
			}
			return n2
		}

		someStr = func(s1, s2 string, both func(s1, s2 string) (string, error)) (string, error) {
			switch {
			case s1 == "" && s2 == "":
				return "", nil
			case s1 != "" && s2 == "":
				return s1, nil
			case s1 == "" && s2 != "":
				return s2, nil
			default:
				return both(s1, s2)
			}
		}

		someNum = func(n1, n2 jsonschema.Num, both func(n1, n2 jx.Num) jx.Num) jsonschema.Num {
			switch {
			case len(n1) == 0 && len(n2) == 0:
				return jsonschema.Num{}
			case len(n1) != 0 && len(n2) == 0:
				return n1
			case len(n1) == 0 && len(n2) != 0:
				return n2
			default:
				if jx.Num(n1).Equal(jx.Num(n2)) {
					return n1
				}
				return jsonschema.Num(both(jx.Num(n1), jx.Num(n2)))
			}
		}

		maxNum = func(n1, n2 jx.Num) jx.Num {
			f1, err := n1.Float64()
			if err != nil {
				panic("unreachable")
			}
			f2, err := n2.Float64()
			if err != nil {
				panic("unreachable")
			}
			if f1 > f2 {
				return n1
			}
			return n2
		}

		minNum = func(n1, n2 jx.Num) jx.Num {
			f1, err := n1.Float64()
			if err != nil {
				panic("unreachable")
			}
			f2, err := n2.Float64()
			if err != nil {
				panic("unreachable")
			}
			if f1 < f2 {
				return n1
			}
			return n2
		}
	)

	switch {
	case s1 == nil && s2 == nil:
		return nil, nil
	case s1 != nil && s2 == nil:
		return s1, nil
	case s1 == nil && s2 != nil:
		return s2, nil
	}

	if allOf := s1.AllOf; len(allOf) > 0 {
		s1, err = mergeNSchemes(allOf)
		if err != nil {
			return nil, errors.Wrap(err, "merge subschemas")
		}
	}
	if allOf := s2.AllOf; len(allOf) > 0 {
		s2, err = mergeNSchemes(allOf)
		if err != nil {
			return nil, errors.Wrap(err, "merge subschemas")
		}
	}

	containsValidators := func(s *jsonschema.Schema) bool {
		if s.Type != "" || s.Format != "" || s.Nullable || len(s.Enum) > 0 || s.DefaultSet || s.ConstSet {
			return true
		}
		if s.Item != nil ||
			s.AdditionalProperties != nil ||
			len(s.PatternProperties) > 0 ||
			len(s.Properties) > 0 ||
			len(s.Required) > 0 {
			return true
		}
		if len(s.OneOf) > 0 || len(s.AnyOf) > 0 || len(s.AllOf) > 0 {
			return true
		}
		if s.Discriminator != nil || s.XML != nil {
			return true
		}
		if len(s.Maximum) > 0 || len(s.Minimum) > 0 || len(s.MultipleOf) > 0 ||
			s.ExclusiveMinimum || s.ExclusiveMaximum {
			return true
		}
		if s.MaxLength != nil || s.MinLength != nil || len(s.Pattern) > 0 {
			return true
		}
		if s.MaxItems != nil || s.MinItems != nil || s.UniqueItems {
			return true
		}
		if s.MaxProperties != nil || s.MinProperties != nil {
			return true
		}
		return false
	}

	switch a, b := containsValidators(s1), containsValidators(s2); [2]bool{a, b} {
	case [2]bool{true, true}, [2]bool{false, false}:
	case [2]bool{true, false}:
		return shallowSchemaCopy(s1), nil
	case [2]bool{false, true}:
		return shallowSchemaCopy(s2), nil
	}

	r := &jsonschema.Schema{
		Format:      s1.Format,
		Nullable:    s1.Nullable || s2.Nullable,
		Description: "Merged schema", // TODO(tdakkota): handle in a better way.
	}

	// Type
	{
		typ, err := someStr(string(s1.Type), string(s2.Type), func(s1, s2 string) (string, error) {
			if s1 == s2 {
				return s1, nil
			}
			return "", errors.Errorf("schema type mismatch: %s and %s", s1, s2)
		})
		if err != nil {
			return nil, err
		}

		r.Type = jsonschema.SchemaType(typ)
	}

	// Format
	{
		format, err := someStr(s1.Format, s2.Format, func(s1, s2 string) (string, error) {
			if s1 == s2 {
				return s1, nil
			}
			return "", errors.Errorf("schema format mismatch: %s and %s", s1, s2)
		})
		if err != nil {
			return nil, err
		}

		r.Format = format
	}

	// Enum
	r.Enum, err = mergeEnums(s1, s2)
	if err != nil {
		return nil, errors.Wrap(err, "enum")
	}

	// Default
	switch {
	case !s1.DefaultSet && !s2.DefaultSet:
		// Nothing to do.
	case s1.DefaultSet && !s2.DefaultSet:
		r.Default = s1.Default
		r.DefaultSet = true
	case !s1.DefaultSet && s2.DefaultSet:
		r.Default = s2.Default
		r.DefaultSet = true
	case s1.DefaultSet && s2.DefaultSet:
		if !reflect.DeepEqual(s1.Default, s2.Default) {
			return nil, errors.New("schemes have different defaults")
		}

		r.Default = s1.Default
		r.DefaultSet = true
	}

	// Const
	switch {
	case s1.ConstSet && !s2.ConstSet:
		r.Const = s1.Const
		r.ConstSet = true
	case !s1.ConstSet && s2.ConstSet:
		r.Const = s2.Const
		r.ConstSet = true
	case s1.ConstSet && s2.ConstSet:
		if !reflect.DeepEqual(s1.Const, s2.Const) {
			return nil, errors.New("schemes have different const values")
		}

		r.Const = s1.Const
		r.ConstSet = true
	}

	// Discriminator
	switch d1, d2 := s1.Discriminator, s2.Discriminator; {
	case d1 != nil && d2 != nil:
		return nil, &ErrNotImplemented{"merge discriminator"} // TODO(tdakkota): implement
	case d1 != nil:
		r.Discriminator = d1
	case d2 != nil:
		r.Discriminator = d2
	}

	// String validation
	{
		r.MaxLength = someU64(s1.MaxLength, s2.MaxLength, selectMinU64)
		r.MinLength = someU64(s1.MinLength, s2.MinLength, selectMaxU64)
		r.Pattern, err = someStr(s1.Pattern, s2.Pattern, func(s1, s2 string) (string, error) {
			if s1 == s2 {
				return s1, nil
			}
			return "", errors.Errorf("cannot merge different patterns: %q and %q", s1, s2)
		})
		if err != nil {
			return nil, errors.Wrap(err, "pattern")
		}
	}

	// Integer, Number validation
	{
		r.Maximum = someNum(s1.Maximum, s2.Maximum, minNum)
		s1.ExclusiveMaximum = s1.ExclusiveMaximum || s2.ExclusiveMaximum

		r.Minimum = someNum(s1.Minimum, s2.Minimum, maxNum)
		r.ExclusiveMinimum = s1.ExclusiveMinimum || s2.ExclusiveMinimum

		// NOTE: We need to refactor ir.Validators to support multiple 'multipleOf's.
		//
		// Most likely it will require rewriting this schema merging code, because
		// we cannot set multiple 'multipleOf's into single jsonschema.Schema.
		// We need to generate ir.Type for each schema in 'allOf' and then merge
		// them into single *ir.Type with all the validation.
		if !bytes.Equal(s1.MultipleOf, s2.MultipleOf) {
			return nil, errors.Errorf("multipleOf is different: %s and %s", s1.MultipleOf, s2.MultipleOf)
		}
		r.MultipleOf = s1.MultipleOf
	}

	// Array validation
	{
		switch {
		case len(s1.Items) > 0 && len(s2.Items) > 0:
			if len(s1.Items) != len(s2.Items) {
				return nil, errors.Errorf("items length is different: %d and %d", len(s1.Items), len(s2.Items))
			}
			result := make([]*jsonschema.Schema, len(s1.Items))
			for i, e1 := range s1.Items {
				e2 := s2.Items[i]
				result[i], err = mergeSchemes(e1, e2)
				if err != nil {
					return nil, errors.Wrapf(err, "merge items[%d]", i)
				}
			}
		case len(s1.Items) == 0 && len(s2.Items) == 0:
			r.Item, err = mergeSchemes(s1.Item, s2.Item)
			if err != nil {
				return nil, errors.Wrap(err, "merge item schema")
			}

			r.MinItems = someU64(s1.MinItems, s2.MinItems, selectMaxU64)
			r.MaxItems = someU64(s1.MaxItems, s2.MaxItems, selectMinU64)
			r.UniqueItems = s1.UniqueItems || s2.UniqueItems
		default:
			return nil, errors.New("can't merge different types of items")
		}
	}

	// Object validation
	{
		if len(s1.PatternProperties) > 0 || len(s2.PatternProperties) > 0 {
			return nil, &ErrNotImplemented{Name: "allOf with patternProperties"}
		}

		switch {
		case s1.AdditionalProperties == nil && s2.AdditionalProperties == nil:
			// Nothing to do.
		case s1.AdditionalProperties != nil && s2.AdditionalProperties == nil:
			r.AdditionalProperties = s1.AdditionalProperties
			r.Item = s1.Item
		case s1.AdditionalProperties == nil && s2.AdditionalProperties != nil:
			r.AdditionalProperties = s2.AdditionalProperties
			r.Item = s2.Item
		case reflect.DeepEqual(s1.AdditionalProperties, s2.AdditionalProperties):
			r.AdditionalProperties = s1.AdditionalProperties
			r.Item, err = mergeSchemes(s1.Item, s2.Item)
			if err != nil {
				return nil, errors.Wrap(err, "merge additionalProperties schema")
			}
		case s1.AdditionalProperties != nil && s2.AdditionalProperties != nil:
			return nil, &ErrNotImplemented{Name: "allOf additionalProperties merging"}
		}

		r.MinProperties = someU64(s1.MinProperties, s2.MinProperties, selectMaxU64)
		r.MaxProperties = someU64(s1.MaxProperties, s2.MaxProperties, selectMinU64)
		r.Properties, err = mergeProperties(s1, s2)
		if err != nil {
			return nil, errors.Wrap(err, "merge properties")
		}
	}

	// oneOf, anyOf
	mergeSum := func(name string, s1, s2 []*jsonschema.Schema) ([]*jsonschema.Schema, error) {
		switch {
		case len(s1) > 0 && len(s2) > 0:
			return nil, &ErrNotImplemented{Name: fmt.Sprintf("allOf with %s", name)}
		case len(s1) > 0:
			return s1, nil
		case len(s2) > 0:
			return s2, nil
		default:
			return nil, nil
		}
	}
	r.OneOf, err = mergeSum("oneOf", s1.OneOf, s2.OneOf)
	if err != nil {
		return nil, errors.Wrap(err, "merge oneOf")
	}
	r.AnyOf, err = mergeSum("anyOf", s1.AnyOf, s2.AnyOf)
	if err != nil {
		return nil, errors.Wrap(err, "merge anyOf")
	}

	return r, nil
}

// mergeProperties finds properties with identical names
// and tries to merge them into one, avoiding duplicates.
func mergeProperties(s1, s2 *jsonschema.Schema) ([]jsonschema.Property, error) {
	var (
		p1 = s1.Properties
		p2 = s2.Properties

		propmap    = make(map[string]jsonschema.Property, len(p1)+len(p2))
		order      = make(map[string]int, len(p1)+len(p2))
		required   = make(map[string]struct{}, len(s1.Required)+len(s2.Required))
		orderIndex = 0
	)
	for _, prop := range s1.Required {
		required[prop] = struct{}{}
	}
	for _, prop := range s2.Required {
		required[prop] = struct{}{}
	}

	// Fill the map with p1 props.
	for _, p := range p1 {
		propmap[p.Name] = p
		order[p.Name] = orderIndex
		orderIndex++
	}

	// Try to merge p2 props.
	for _, p := range p2 {
		if confP, ok := propmap[p.Name]; ok {
			// Property name conflict.
			s, err := mergeSchemes(p.Schema, confP.Schema)
			if err != nil {
				return nil, errors.Wrap(err, "try to merge conflicting property schemas")
			}

			propmap[p.Name] = jsonschema.Property{
				Name:        p.Name,
				Description: "Merged property", // TODO(tdakkota): handle in a better way.
				Schema:      s,
				Required:    p.Required || confP.Required,
			}
			continue
		}

		propmap[p.Name] = p
		order[p.Name] = orderIndex
		orderIndex++
	}

	result := make([]jsonschema.Property, len(propmap))
	for name, p := range propmap {
		_, require := required[p.Name]
		p.Required = p.Required || require
		result[order[name]] = p
	}

	return result, nil
}

func mergeEnums(s1, s2 *jsonschema.Schema) ([]any, error) {
	switch {
	case len(s1.Enum) == 0 && len(s2.Enum) == 0:
		return nil, nil
	case len(s1.Enum) > 0 && len(s2.Enum) == 0:
		return s1.Enum, nil
	case len(s1.Enum) == 0 && len(s2.Enum) > 0:
		return s2.Enum, nil
	}

	var (
		small = s1.Enum
		big   = s2.Enum
	)
	if len(s1.Enum) > len(s2.Enum) {
		small = s2.Enum
		big = s1.Enum
	}
	// Keep values that are present in both enums.
	var result []any
	for _, v := range small {
		// FIXME(tdakkota): quadratic complexity.
		if slices.ContainsFunc(big, func(x any) bool {
			return reflect.DeepEqual(x, v)
		}) {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return nil, &ErrNotImplemented{Name: "allOf enum merging"}
	}
	return result, nil
}
