package gen

import (
	"cmp"
	"fmt"
	"path"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/naming"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

type refNamer func(ref jsonschema.Ref) (string, error)

const defaultSchemaDepthLimit = 1000

type schemaGen struct {
	side      []*ir.Type
	localRefs map[jsonschema.Ref]*ir.Type
	lookupRef func(ref jsonschema.Ref) (*ir.Type, bool)
	nameRef   func(ref jsonschema.Ref) (string, error)
	fieldMut  func(*ir.Field) error
	fail      func(err error) error
	imports   map[string]string

	depthLimit int
	depthCount int

	request bool // true if generating for request body

	log *zap.Logger
}

func newSchemaGen(lookupRef func(ref jsonschema.Ref) (*ir.Type, bool)) *schemaGen {
	return &schemaGen{
		localRefs: map[jsonschema.Ref]*ir.Type{},
		lookupRef: lookupRef,
		nameRef: func(ref jsonschema.Ref) (string, error) {
			name, err := pascal(cleanRef(ref))
			if err != nil {
				return "", err
			}
			return name, nil
		},
		fail: func(err error) error {
			return err
		},
		imports:    defaultImports(),
		depthLimit: defaultSchemaDepthLimit,
		log:        zap.NewNop(),
	}
}

func variantFieldName(t *ir.Type) string {
	return naming.Capitalize(t.NamePostfix())
}

type schemaDepthError struct {
	limit int
}

func (e *schemaDepthError) Error() string {
	return fmt.Sprintf("schema depth limit (%d) exceeded", e.limit)
}

func handleSchemaDepth(s *jsonschema.Schema, rerr *error) {
	r := recover()
	if r == nil {
		return
	}

	e, ok := r.(*schemaDepthError)
	if !ok {
		panic(r)
	}
	*rerr = e

	// Ensure that schema is not nil.
	if s == nil {
		return
	}
	ptr := s.Pointer

	// Try to use location.Error.
	pos, ok := ptr.Position()
	if !ok {
		return
	}
	*rerr = &location.Error{
		File: ptr.File(),
		Pos:  pos,
		Err:  e,
	}
}

func (g *schemaGen) generate(name string, schema *jsonschema.Schema, optional bool) (*ir.Type, error) {
	g.depthCount++
	if g.depthCount > g.depthLimit {
		// Panicing is not cool, but is better rather than wrap the error N = depthLimit times.
		panic(&schemaDepthError{limit: g.depthLimit})
	}
	defer func() {
		g.depthCount--
	}()

	schema = transformSchema(schema)

	t, err := g.generate2(name, schema)
	if err != nil {
		return nil, err
	}

	nullable := schema != nil && schema.Nullable
	t, err = boxType(t, ir.GenericVariant{
		Optional: optional,
		Nullable: nullable,
	})
	if err != nil {
		return nil, err
	}

	if t.IsGeneric() {
		g.side = append(g.side, t)
	}

	// TODO: update refcache to point to new boxed type?
	return t, nil
}

func (g *schemaGen) generate2(name string, schema *jsonschema.Schema) (ret *ir.Type, err error) {
	if schema == nil {
		// Empty schema (no schema field in OpenAPI spec).
		// For responses: Allow jx.Raw since client must handle unknown JSON from server.
		// For requests: Reject to avoid clients sending arbitrary data without spec guidance.
		if g.request {
			return nil, &ErrNotImplemented{Name: "empty schema in request body"}
		}
		// For responses, treat as "any valid JSON value" (jx.Raw).
		// Consistent with array item handling (line 437).
		return ir.Any(nil), nil
	}

	if ref := schema.Ref; !ref.IsZero() {
		if t, ok := g.lookupRef(ref); ok {
			return t, nil
		}
		if t, ok := g.localRefs[ref]; ok {
			return t, nil
		}

		name, err = g.nameRef(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "schema name: %q", ref)
		}
	}

	if schema.XOgenType != "" {
		t, err := ir.External(schema)
		if err != nil {
			return nil, errors.Wrap(err, "external type")
		}

		if pkgPath := t.External.PackagePath; pkgPath != "" {
			if alias, ok := g.imports[pkgPath]; ok {
				t.External.ImportAlias = alias
			} else {
				aliases := make(map[string]struct{}, len(g.imports))
				for k, v := range g.imports {
					aliases[cmp.Or(v, path.Base(k))] = struct{}{}
				}
				pkgName := t.External.PackageName
				if _, ok := aliases[pkgName]; ok {
					for i := 2; true; i++ {
						t.External.ImportAlias = fmt.Sprintf("%s%d", pkgName, i)
						if _, ok := aliases[t.External.ImportAlias]; !ok {
							break
						}
					}
				}
			}
			t.Primitive = t.External.Primitive()
			g.imports[pkgPath] = t.External.ImportAlias
		}

		return g.regtype(name, t), nil
	}

	if schema.DefaultSet {
		var implErr error
		switch {
		case schema.Type == jsonschema.Object:
			implErr = &ErrNotImplemented{Name: "object defaults"}
		case schema.Type == jsonschema.Array:
			implErr = &ErrNotImplemented{Name: "array defaults"}
		case schema.Type == jsonschema.Empty ||
			len(schema.AnyOf)+len(schema.OneOf) > 0:
			implErr = &ErrNotImplemented{Name: "complex defaults"}
		}
		// Do not fail schema generation if we cannot handle defaults.
		if err := g.fail(implErr); err != nil {
			return nil, err
		}

		if implErr == nil {
			if err := g.checkDefaultType(schema, schema.Default); err != nil {
				return nil, errors.Wrap(err, "check default type")
			}
		}
		schema.DefaultSet = implErr == nil
	}

	if schema.UniqueItems {
		item := schema.Item
		// Only error on truly invalid cases (nil or empty item type).
		// Complex types (Array, Object) are now supported via Equal/Hash generation.
		if item == nil || item.Type == "" {
			return nil, &ErrNotImplemented{Name: "empty uniqueItems"}
		}
	}

	if n := schema.XOgenName; n != "" {
		name = n
	} else if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "R" + name
	}

	var (
		oneOf                   *ir.Type
		anyOf                   *ir.Type
		checkOnlyObjectVariants = func(sum []*jsonschema.Schema) error {
			for _, s := range sum {
				if s.Type == jsonschema.Object {
					continue
				}

				ptr := s.Pointer
				err := errors.Errorf("can't merge object with %s", s.Type)

				pos, ok := ptr.Position()
				if !ok {
					return err
				}

				return &location.Error{
					File: ptr.File(),
					Pos:  pos,
					Err:  err,
				}
			}
			return nil
		}
	)
	switch {
	case len(schema.AnyOf) > 0:
		side := schema.Type == jsonschema.Object
		sumName := name
		if side {
			sumName += "Sum"
		}
		t, err := g.anyOf(sumName, schema, side)
		if err != nil {
			return nil, errors.Wrap(errors.Wrap(err, "anyOf"), sumName)
		}

		if !side {
			return t, nil
		}
		if err := checkOnlyObjectVariants(schema.AnyOf); err != nil {
			return nil, err
		}
		anyOf = t
	case len(schema.AllOf) > 0:
		t, err := g.allOf(name, schema)
		if err != nil {
			return nil, errors.Wrap(errors.Wrap(err, "allOf"), name)
		}
		return t, nil
	case len(schema.OneOf) > 0:
		side := schema.Type == jsonschema.Object
		sumName := name
		if side {
			sumName += "Sum"
		}
		t, err := g.oneOf(sumName, schema, side)
		if err != nil {
			return nil, errors.Wrap(errors.Wrap(err, "oneOf"), sumName)
		}

		if !side {
			return t, nil
		}
		if err := checkOnlyObjectVariants(schema.OneOf); err != nil {
			return nil, err
		}
		oneOf = t
	case len(schema.Enum) > 0:
		switch schema.Type {
		case jsonschema.String,
			jsonschema.Integer,
			jsonschema.Number,
			jsonschema.Boolean,
			jsonschema.Null:
			// Primitive enums are handled below
		case jsonschema.Object:
			// Non-primitive object enums generate sum types with struct variants.
			// Each enum value becomes a concrete struct type.
			t, err := g.nonPrimitiveObjectEnum(name, schema)
			if err != nil {
				return nil, errors.Wrap(err, "non-primitive object enum")
			}
			return t, nil
		case jsonschema.Array, jsonschema.Empty:
			// Array enums and empty type enums are treated as "any" type.
			// The enum constraint is documented in OpenAPI but not enforced at runtime.
			return g.regtype(name, ir.Any(schema)), nil
		}
	}

	switch schema.Type {
	case jsonschema.Object:
		kind := ir.KindStruct

		hasProps := len(schema.Properties) > 0
		hasAdditionalProps := false
		denyAdditionalProps := false
		if p := schema.AdditionalProperties; p != nil {
			hasAdditionalProps = *p
			denyAdditionalProps = !*p
		}
		hasPatternProps := len(schema.PatternProperties) > 0
		isPatternSingle := len(schema.PatternProperties) == 1

		if !hasProps {
			if (!hasAdditionalProps && isPatternSingle) || (hasAdditionalProps && !hasPatternProps) {
				kind = ir.KindMap
			}
		}

		s := g.regtype(name, &ir.Type{
			Kind:                kind,
			Name:                name,
			Schema:              schema,
			DenyAdditionalProps: denyAdditionalProps,
		})
		s.Validators.SetObject(schema)
		s.Validators.SetOgenValidate(schema)

		type fieldSlot struct {
			// Stores spec name of the field.
			original      string
			nameDefinedAt location.Pointer
		}
		fieldNames := map[string]fieldSlot{}

		addField := func(f *ir.Field, adding fieldSlot) error {
			if m := g.fieldMut; m != nil {
				if err := m(f); err != nil {
					return err
				}
			}
			existing, ok := fieldNames[f.Name]
			if !ok {
				s.Fields = append(s.Fields, f)
				fieldNames[f.Name] = adding
				return nil
			}

			err := errors.Errorf("conflict: field %q already defined by %s", f.Name, existing.original)
			ptr := adding.nameDefinedAt
			if _, ok := ptr.Position(); !ok {
				// Use existing as a fallback.
				ptr = existing.nameDefinedAt
			}

			pos, ok := ptr.Position()
			if !ok {
				return err
			}

			return &location.Error{
				File: ptr.File(),
				Pos:  pos,
				Err:  err,
			}
		}

		for i := range schema.Properties {
			prop := schema.Properties[i]
			propTypeName, err := pascalSpecial(name, prop.Name)
			if err != nil {
				return nil, errors.Wrapf(err, "property type name: %q", prop.Name)
			}

			t, err := g.generate(propTypeName, prop.Schema, !prop.Required)
			if err != nil {
				return nil, errors.Wrapf(err, "field %s", prop.Name)
			}

			var (
				fieldName string
				slot      fieldSlot
			)
			if n := prop.X.Name; n != nil {
				fieldName = *n

				slot = fieldSlot{
					original:      fmt.Sprintf("property %q (overridden by extension as %q)", prop.Name, *n),
					nameDefinedAt: prop.X.Field("name"),
				}
			} else {
				propertyName := strings.TrimSpace(prop.Name)
				if propertyName == "" {
					propertyName = fmt.Sprintf("Field%d", i)
				}

				generated, err := pascalSpecial(propertyName)
				if err != nil {
					return nil, errors.Wrapf(err, "property name: %q", propertyName)
				}
				fieldName = generated

				slot = fieldSlot{
					original:      fmt.Sprintf("property %q", prop.Name),
					nameDefinedAt: schema.Pointer.Field("properties").Key(prop.Name),
				}
			}

			if err := addField(&ir.Field{
				Name: fieldName,
				Type: t,
				Tag: ir.Tag{
					JSON:      prop.Name,
					ExtraTags: prop.Schema.ExtraTags,
				},
				Spec: &prop,
			}, slot); err != nil {
				return nil, err
			}
		}

		item := func(prefix string, schItem *jsonschema.Schema) (*ir.Type, error) {
			if schItem == nil {
				return ir.Any(schItem), nil
			}
			return g.generate(prefix+"Item", schItem, false)
		}

		if hasAdditionalProps {
			mapType := s
			// Create special field for additionalProperties.
			if s.Kind != ir.KindMap {
				mapType = g.regtype(name, &ir.Type{
					Kind: ir.KindMap,
					Name: s.Name + "Additional",
				})

				const key = "additionalProperties"
				slot := fieldSlot{
					original:      key,
					nameDefinedAt: schema.Key(key),
				}
				if err := addField(&ir.Field{
					Name:   "AdditionalProps",
					Type:   mapType,
					Inline: ir.InlineAdditional,
				}, slot); err != nil {
					return nil, err
				}
			}

			mapType.Item, err = item(mapType.Name, schema.Item)
			if err != nil {
				return nil, errors.Wrap(err, "item")
			}
		}
		if hasPatternProps {
			if s.Kind == ir.KindMap {
				pp := schema.PatternProperties[0]
				s.MapPattern = pp.Pattern
				s.Item, err = item(s.Name, pp.Schema)
				if err != nil {
					return nil, errors.Wrapf(err, "pattern schema %q", s.MapPattern)
				}
			} else {
				for idx, pp := range schema.PatternProperties {
					suffix := fmt.Sprintf("Pattern%d", idx)

					mapType := g.regtype(name, &ir.Type{
						Kind:       ir.KindMap,
						Name:       s.Name + suffix,
						MapPattern: pp.Pattern,
					})
					mapType.Item, err = item(mapType.Name, pp.Schema)
					if err != nil {
						return nil, errors.Wrapf(err, "pattern schema [%d] %q", idx, pp.Pattern)
					}

					fieldName := suffix + "Props"
					slot := fieldSlot{
						original:      fmt.Sprintf("pattern %q", pp.Pattern),
						nameDefinedAt: pp.Schema.Pointer,
					}
					if err := addField(&ir.Field{
						Name:   fieldName,
						Type:   mapType,
						Inline: ir.InlinePattern,
					}, slot); err != nil {
						return nil, err
					}
				}
			}
		}
		if anyOf != nil {
			slot := fieldSlot{
				original:      "anyOf",
				nameDefinedAt: schema.Key("anyOf"),
			}
			if err := addField(&ir.Field{
				Name:   "AnyOf",
				Type:   anyOf,
				Inline: ir.InlineSum,
			}, slot); err != nil {
				return nil, err
			}
		}
		if oneOf != nil {
			slot := fieldSlot{
				original:      "oneOf",
				nameDefinedAt: schema.Key("oneOf"),
			}
			if err := addField(&ir.Field{
				Name:   "OneOf",
				Type:   oneOf,
				Inline: ir.InlineSum,
			}, slot); err != nil {
				return nil, err
			}
		}

		return s, nil
	case jsonschema.Array:
		if tuple := schema.Items; len(tuple) > 0 {
			ret := g.regtype(name, &ir.Type{
				Kind:   ir.KindStruct,
				Name:   name,
				Schema: schema,
				Tuple:  true,
			})

			for i, item := range tuple {
				fieldName := fmt.Sprintf("V%d", i)
				if item.XOgenName != "" {
					// Using the name from the ogen schema extension.
					// Avoiding name conflicts is up to user.
					fieldName = item.XOgenName
				}
				f, err := g.generate(name+fieldName, item, false)
				if err != nil {
					return nil, errors.Wrapf(err, "tuple element %d", i)
				}
				ret.Fields = append(ret.Fields, &ir.Field{
					Name: fieldName,
					Type: f,
				})
			}

			return ret, nil
		}
		array := &ir.Type{
			Kind:        ir.KindArray,
			Schema:      schema,
			NilSemantic: ir.NilInvalid,
		}
		array.Validators.SetArray(schema)
		array.Validators.SetOgenValidate(schema)

		ret := g.regtype(name, array)
		if item := schema.Item; item != nil {
			array.Item, err = g.generate(name+"Item", item, false)
			if err != nil {
				return nil, errors.Wrap(err, "item")
			}
		} else {
			array.Item = ir.Any(item)
		}

		return ret, nil

	case jsonschema.String, jsonschema.Integer, jsonschema.Number, jsonschema.Boolean, jsonschema.Null:
		t, err := g.primitive(name, schema)
		if err != nil {
			return nil, errors.Wrap(err, "primitive")
		}

		fields := []zap.Field{
			zapPosition(schema),
			zap.String("type", schema.Type.String()),
			zap.String("format", schema.Format),
			zap.String("go_type", t.Go()),
		}
		switch schema.Type {
		case jsonschema.String:
			if err := t.Validators.SetString(schema); err != nil {
				return nil, errors.Wrap(err, "string validator")
			}
			if t.Validators.String.Set() {
				switch t.Primitive {
				case ir.String, ir.ByteSlice:
				default:
					g.log.Warn("String validator cannot be applied to generated type and will be ignored", fields...)
				}
			}
		case jsonschema.Integer:
			if err := t.Validators.SetInt(schema); err != nil {
				return nil, errors.Wrap(err, "int validator")
			}
			if t.Validators.Int.Set() && !t.IsInteger() {
				g.log.Warn("Int validator cannot be applied to generated type and will be ignored", fields...)
			}
		case jsonschema.Number:
			if t.IsDecimal() {
				if err := t.Validators.SetDecimal(schema); err != nil {
					return nil, errors.Wrap(err, "decimal validator")
				}
			} else {
				if err := t.Validators.SetFloat(schema); err != nil {
					return nil, errors.Wrap(err, "float validator")
				}
				if t.Validators.Float.Set() && !t.IsFloat() {
					g.log.Warn("Float validator cannot be applied to generated type and will be ignored", fields...)
				}
			}
		}
		t.Validators.SetOgenValidate(schema)

		return g.regtype(name, t), nil
	case jsonschema.Empty:
		g.log.Info("Type is not defined, using any",
			zapPosition(schema),
			zap.String("name", name),
		)
		return g.regtype(name, ir.Any(schema)), nil
	default:
		panic(unreachable(schema.Type))
	}
}

func (g *schemaGen) regtype(name string, t *ir.Type) *ir.Type {
	if t.Schema != nil {
		if ref := t.Schema.Ref; !ref.IsZero() {
			if t.Is(ir.KindPrimitive, ir.KindArray, ir.KindAny) {
				t = ir.Alias(name, t)
			}

			g.localRefs[ref] = t
			return t
		}
	}

	if t.Is(ir.KindStruct, ir.KindMap, ir.KindEnum, ir.KindSum) {
		g.side = append(g.side, t)
	}

	return t
}

func (g *schemaGen) checkDefaultType(s *jsonschema.Schema, val any) error {
	if s == nil {
		// Schema has no validators.
		return nil
	}
	if val == nil && s.Nullable {
		return nil
	}

	var ok bool
	switch s.Type {
	case jsonschema.Object:
		_, ok = val.(map[string]any)
	case jsonschema.Array:
		_, ok = val.([]any)
	case jsonschema.Integer:
		_, ok = val.(int64)
	case jsonschema.Number:
		_, ok = val.(int64)
		if !ok {
			_, ok = val.(float64)
		}
	case jsonschema.String:
		_, ok = val.(string)
	case jsonschema.Boolean:
		_, ok = val.(bool)
	case jsonschema.Null:
		ok = val == nil
	}

	if !ok {
		err := errors.Errorf("expected schema type is %q, default value is %T", s.Type, val)
		p := s.Field("default")

		pos, ok := p.Position()
		if !ok {
			return err
		}

		return &location.Error{
			File: p.File(),
			Pos:  pos,
			Err:  err,
		}
	}

	return nil
}

// nonPrimitiveObjectEnum generates a sum type for object enums.
// Each enum value becomes a concrete struct variant.
func (g *schemaGen) nonPrimitiveObjectEnum(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	if len(schema.Enum) == 0 {
		return nil, errors.New("enum has no values")
	}

	// Convert enum values to map[string]any
	enumObjects := make([]map[string]any, 0, len(schema.Enum))
	for i, v := range schema.Enum {
		obj, ok := v.(map[string]any)
		if !ok {
			return nil, errors.Errorf("enum[%d]: expected object, got %T", i, v)
		}
		enumObjects = append(enumObjects, obj)
	}

	// Find a discriminating field - a string field with unique values across all variants
	discriminatorField, variantNames := findEnumDiscriminator(enumObjects)
	if discriminatorField == "" {
		// No discriminator found, fall back to index-based naming
		for i := range enumObjects {
			variantNames = append(variantNames, fmt.Sprintf("Variant%d", i))
		}
	}

	// Create the sum type
	sum := g.regtype(name, &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: schema,
	})

	// Generate struct types for each enum value
	variants := make([]*ir.Type, 0, len(enumObjects))
	for i, obj := range enumObjects {
		variantName := name + variantNames[i]
		variantSchema := inferSchemaFromObject(obj)

		// Generate the variant struct type
		variantType := &ir.Type{
			Kind:   ir.KindStruct,
			Name:   variantName,
			Schema: variantSchema,
		}

		// Add fields from the object
		for fieldName, fieldValue := range obj {
			fieldSchema := inferSchemaFromValue(fieldValue)
			fieldType := g.inferTypeFromValue(fieldValue, fieldSchema)

			field := &ir.Field{
				Name: naming.Capitalize(fieldName),
				Type: fieldType,
				Tag: ir.Tag{
					JSON: fieldName,
				},
				Spec: &jsonschema.Property{
					Name:     fieldName,
					Schema:   fieldSchema,
					Required: true,
				},
			}
			variantType.Fields = append(variantType.Fields, field)
		}

		// Register and add to variants
		g.regtype(variantName, variantType)
		variants = append(variants, variantType)
	}

	sum.SumOf = variants

	// Set up discrimination
	if discriminatorField != "" {
		// Value-based discrimination on the discriminator field
		valueToVariant := make(map[string]string)
		for i, obj := range enumObjects {
			if val, ok := obj[discriminatorField].(string); ok {
				valueToVariant[val] = variants[i].Name + name
			}
		}
		sum.SumSpec.ValueDiscriminators = map[string]ir.ValueDiscriminator{
			discriminatorField: {
				FieldName:      discriminatorField,
				ValueToVariant: valueToVariant,
			},
		}
	} else {
		// No discriminator field found, use type-based discrimination as fallback
		sum.SumSpec.TypeDiscriminator = true
	}

	return sum, nil
}

// findEnumDiscriminator finds a string field that has unique values across all enum objects.
func findEnumDiscriminator(objects []map[string]any) (string, []string) {
	if len(objects) == 0 {
		return "", nil
	}

	// Find all string fields present in all objects
	stringFields := make(map[string][]string)
	for _, obj := range objects {
		for k, v := range obj {
			if s, ok := v.(string); ok {
				stringFields[k] = append(stringFields[k], s)
			}
		}
	}

	// Find a field with unique values across all objects
	for field, values := range stringFields {
		if len(values) != len(objects) {
			continue // Field not present in all objects
		}

		// Check if all values are unique
		seen := make(map[string]bool)
		allUnique := true
		for _, v := range values {
			if seen[v] {
				allUnique = false
				break
			}
			seen[v] = true
		}

		if allUnique {
			// Use these values as variant names (capitalized)
			variantNames := make([]string, len(values))
			for i, v := range values {
				variantNames[i] = naming.Capitalize(v)
			}
			return field, variantNames
		}
	}

	return "", nil
}

// inferSchemaFromObject creates a jsonschema.Schema from an object literal.
func inferSchemaFromObject(obj map[string]any) *jsonschema.Schema {
	schema := &jsonschema.Schema{
		Type: jsonschema.Object,
	}
	for fieldName, fieldValue := range obj {
		prop := jsonschema.Property{
			Name:     fieldName,
			Schema:   inferSchemaFromValue(fieldValue),
			Required: true,
		}
		schema.Properties = append(schema.Properties, prop)
	}
	return schema
}

// inferSchemaFromValue creates a jsonschema.Schema from a JSON value.
func inferSchemaFromValue(v any) *jsonschema.Schema {
	switch val := v.(type) {
	case string:
		return &jsonschema.Schema{Type: jsonschema.String}
	case int64:
		return &jsonschema.Schema{Type: jsonschema.Integer}
	case float64:
		return &jsonschema.Schema{Type: jsonschema.Number}
	case bool:
		return &jsonschema.Schema{Type: jsonschema.Boolean}
	case nil:
		return &jsonschema.Schema{Type: jsonschema.Null}
	case []any:
		schema := &jsonschema.Schema{Type: jsonschema.Array}
		if len(val) > 0 {
			schema.Item = inferSchemaFromValue(val[0])
		}
		return schema
	case map[string]any:
		return inferSchemaFromObject(val)
	default:
		return &jsonschema.Schema{}
	}
}

// inferTypeFromValue creates an ir.Type from a JSON value.
func (g *schemaGen) inferTypeFromValue(v any, schema *jsonschema.Schema) *ir.Type {
	switch v.(type) {
	case string:
		return ir.Primitive(ir.String, schema)
	case int64:
		return ir.Primitive(ir.Int64, schema)
	case float64:
		return ir.Primitive(ir.Float64, schema)
	case bool:
		return ir.Primitive(ir.Bool, schema)
	case nil:
		return ir.Primitive(ir.Null, schema)
	case []any:
		return ir.Array(g.inferTypeFromValue(nil, nil), ir.NilInvalid, schema)
	case map[string]any:
		return ir.Any(schema)
	default:
		return ir.Any(schema)
	}
}
