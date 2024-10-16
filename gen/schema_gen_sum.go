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
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

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

					pos, ok := s.Pointer.Position()
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
	return nil, &ErrNotImplemented{"complex anyOf"}
}

func (g *schemaGen) oneOf(name string, schema *jsonschema.Schema, side bool) (*ir.Type, error) {
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
		variants, err := g.collectSumVariants(name, schema.OneOf)
		if err != nil {
			return nil, errors.Wrap(err, "collect variants")
		}
		sum.SumOf = variants
	}

	// 1st case: explicit discriminator.
	if d := schema.Discriminator; d != nil {
		propName := schema.Discriminator.PropertyName
		sum.SumSpec.Discriminator = propName
		for k, v := range schema.Discriminator.Mapping {
			// Explicit mapping.
			var found bool
			for i, s := range sum.SumOf {
				if !s.Is(ir.KindStruct, ir.KindMap) {
					return nil, errors.Wrapf(&ErrNotImplemented{"unsupported sum type variant"}, "%q", s.Kind)
				}
				vschema := s.Schema
				if vschema == nil {
					vschema = schema.OneOf[i]
				}

				if vschema.Ref == v.Ref {
					found = true
					sum.SumSpec.Mapping = append(sum.SumSpec.Mapping, ir.SumSpecMap{
						Key:  k,
						Type: s,
					})
				}
			}
			if !found {
				return nil, errors.Errorf("discriminator: unable to map %q to %q", k, v.Ref)
			}
		}
		if len(sum.SumSpec.Mapping) == 0 {
			// Implicit mapping, defaults to type name.
			keys := map[string]struct{}{}
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
						return "", errors.Wrap(
							&ErrNotImplemented{"complicated reference"},
							"unable to extract schema name",
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
				})
			}
		}
		slices.SortStableFunc(sum.SumSpec.Mapping, func(a, b ir.SumSpecMap) int {
			return strings.Compare(a.Key, b.Key)
		})
		return sum, nil
	}

	// 2nd case: distinguish by serialization type.
	if canUseTypeDiscriminator(sum.SumOf, true) {
		sum.SumSpec.TypeDiscriminator = true
		return sum, nil
	}

	// 3rd case: distinguish by unique fields.

	// Determine unique fields for each SumOf variant.
	uniq := map[string]map[string]struct{}{}

	for _, s := range sum.SumOf {
		uniq[s.Name] = map[string]struct{}{}
		if !s.Is(ir.KindStruct) {
			return nil, errors.Wrapf(&ErrNotImplemented{Name: "discriminator inference"},
				"oneOf %s: variant %s: no unique fields, "+
					"unable to parse without discriminator", sum.Name, s.Name,
			)
		}
		for _, f := range s.JSON().Fields() {
			uniq[s.Name][f.Tag.JSON] = struct{}{}
		}
	}
	{
		// Collect fields that common for at least 2 variants.
		commonFields := map[string]struct{}{}
		for _, variant := range sum.SumOf {
			k := variant.Name
			fields := uniq[k]
			for _, otherVariant := range sum.SumOf {
				otherK := otherVariant.Name
				if otherK == k {
					continue
				}
				otherFields := uniq[otherK]
				for otherField := range otherFields {
					if _, has := fields[otherField]; has {
						// variant and otherVariant have common field otherField.
						commonFields[otherField] = struct{}{}
					}
				}
			}
		}

		// Delete such fields.
		for field := range commonFields {
			for _, variant := range sum.SumOf {
				delete(uniq[variant.Name], field)
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
	var variants []sumVariant
	for _, s := range sum.SumOf {
		k := s.Name
		f := uniq[k]
		variants = append(variants, sumVariant{
			Name:   k,
			Unique: xmaps.SortedKeys(f),
		})
	}
	slices.SortStableFunc(variants, func(a, b sumVariant) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, v := range variants {
		for _, s := range sum.SumOf {
			if s.Name != v.Name {
				continue
			}
			if len(s.SumSpec.Unique) > 0 {
				continue
			}
			for _, f := range s.JSON().Fields() {
				if !slices.Contains(v.Unique, f.Tag.JSON) {
					continue
				}
				s.SumSpec.Unique = append(s.SumSpec.Unique, f)
			}
		}
	}
	return sum, nil
}

func (g *schemaGen) allOf(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	if err := ensureNoInfiniteRecursion(schema); err != nil {
		return nil, err
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
		if s.Type != "" || s.Format != "" || s.Nullable || len(s.Enum) > 0 || s.DefaultSet {
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
