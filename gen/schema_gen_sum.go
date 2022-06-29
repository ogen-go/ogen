package gen

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"sort"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func canUseTypeDiscriminator(sum []*ir.Type) bool {
	var (
		// Collect map of variant kinds.
		typeMap      = map[string]struct{}{}
		collectTypes func(sum []*ir.Type)
	)
	collectTypes = func(sum []*ir.Type) {
		for _, variant := range sum {
			typ := variant.JSON().Type()
			if typ == "" {
				if variant.IsSum() {
					collectTypes(variant.SumOf)
				}
				continue
			}
			typeMap[typ] = struct{}{}
		}
	}

	for _, s := range sum {
		typ := s.JSON().Type()
		switch {
		case s.IsSum():
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
			// Type kind is not unique, so we can distinguish variants by type.
			return false
		}
		typeMap[typ] = struct{}{}
	}
	return true
}

func ensureNoInfiniteRecursion(parent *jsonschema.Schema) error {
	var do func(map[string]struct{}, []*jsonschema.Schema) error
	do = func(ctx map[string]struct{}, schemas []*jsonschema.Schema) error {
		for i, s := range schemas {
			if ref := s.Ref; ref != "" {
				if _, ok := ctx[ref]; ok {
					return errors.Errorf("reference %q [%d] leads to infinite recursion", ref, i)
				}
				ctx[ref] = struct{}{}
			}
			switch {
			case len(s.OneOf) > 0:
				if err := do(ctx, s.OneOf); err != nil {
					return errors.Wrapf(err, "oneOf %q [%d]", s.Ref, i)
				}
			case len(s.AllOf) > 0:
				if err := do(ctx, s.AllOf); err != nil {
					return errors.Wrapf(err, "allOf %q [%d]", s.Ref, i)
				}
			case len(s.AnyOf) > 0:
				if err := do(ctx, s.AnyOf); err != nil {
					return errors.Wrapf(err, "anyOf %q [%d]", s.Ref, i)
				}
			}
			delete(ctx, s.Ref)
		}
		return nil
	}

	return do(map[string]struct{}{}, []*jsonschema.Schema{parent})
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

	names := map[string]struct{}{}
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

func (g *schemaGen) anyOf(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	if err := ensureNoInfiniteRecursion(schema); err != nil {
		return nil, err
	}

	sum := g.regtype(name, &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: schema,
	})
	{
		variants, err := g.collectSumVariants(name, schema.AnyOf)
		if err != nil {
			return nil, errors.Wrap(err, "collect variants")
		}
		sum.SumOf = variants
	}

	// Here we try to create sum type from anyOf for variants with JSON type-based discriminator.
	if canUseTypeDiscriminator(sum.SumOf) {
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

func (g *schemaGen) oneOf(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	if err := ensureNoInfiniteRecursion(schema); err != nil {
		return nil, err
	}

	sum := g.regtype(name, &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: schema,
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
				var ref string
				if s.Schema != nil {
					ref = s.Schema.Ref
				} else {
					ref = schema.OneOf[i].Ref
				}

				if ref == v || path.Base(ref) == v {
					found = true
					sum.SumSpec.Mapping = append(sum.SumSpec.Mapping, ir.SumSpecMap{
						Key:  k,
						Type: s,
					})

					// Filter discriminator field in-place.
					n := 0
					for _, f := range s.Fields {
						if f.Tag.JSON != propName {
							s.Fields[n] = f
							n++
						}
					}
					s.Fields = s.Fields[:n]
				}
			}
			if !found {
				return nil, errors.Errorf("discriminator: unable to map %q to %q", k, v)
			}
		}
		if len(sum.SumSpec.Mapping) == 0 {
			// Implicit mapping, defaults to type name.
			for i, s := range sum.SumOf {
				var ref string
				if s.Schema != nil {
					ref = s.Schema.Ref
				} else {
					ref = schema.OneOf[i].Ref
				}

				sum.SumSpec.Mapping = append(sum.SumSpec.Mapping, ir.SumSpecMap{
					Key:  path.Base(ref),
					Type: s,
				})
			}
		}
		sort.SliceStable(sum.SumSpec.Mapping, func(i, j int) bool {
			a := sum.SumSpec.Mapping[i]
			b := sum.SumSpec.Mapping[j]
			return a.Key < b.Key
		})
		return sum, nil
	}

	// 2nd case: distinguish by serialization type.
	if canUseTypeDiscriminator(sum.SumOf) {
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
			uniq[s.Name][f.Name] = struct{}{}
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
		metNoUniqueFields := false
		for _, variant := range sum.SumOf {
			k := variant.Name
			if len(uniq[k]) == 0 {
				if metNoUniqueFields {
					// Unable to deterministically select sub-schema only on fields.
					return nil, errors.Wrapf(&ErrNotImplemented{Name: "discriminator inference"},
						"oneOf %s: variant %s: no unique fields, "+
							"unable to parse without discriminator", sum.Name, k,
					)
				}

				// Set mapping without unique fields as default
				sum.SumSpec.DefaultMapping = k
				metNoUniqueFields = true
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
		v := sumVariant{
			Name: k,
		}
		for fieldName := range f {
			v.Unique = append(v.Unique, fieldName)
		}
		sort.Strings(v.Unique)
		variants = append(variants, v)
	}
	sort.SliceStable(variants, func(i, j int) bool {
		a := variants[i]
		b := variants[j]
		return a.Name < b.Name
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
				var skip bool
				for _, n := range v.Unique {
					if n == f.Name {
						skip = true // not unique
						break
					}
				}
				if !skip {
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

	return g.generate(name, mergedSchema, false)
}

func mergeNSchemes(ss []*jsonschema.Schema) (_ *jsonschema.Schema, err error) {
	if len(ss) < 1 {
		panic("unreachable")
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
	if s1 == nil || s2 == nil {
		return nil, errors.Errorf("schema is null or empty")
	}

	if s1.Type != s2.Type {
		return nil, errors.Errorf("schema type mismatch: %s and %s", s1.Type, s2.Type)
	}

	if s1.Format != s2.Format {
		return nil, errors.Errorf("schema format mismatch: %s and %s", s1.Format, s2.Format)
	}

	enum, err := mergeEnums(s1, s2)
	if err != nil {
		return nil, errors.Wrap(err, "enum")
	}

	r := &jsonschema.Schema{
		Type:        s1.Type,
		Format:      s1.Format,
		Enum:        enum,
		Nullable:    s1.Nullable || s2.Nullable,
		Description: "Merged schema",
	}

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
			return nil, errors.Errorf("schemes have different defaults")
		}

		r.Default = s1.Default
		r.DefaultSet = true
	}

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

	// JSONSchema says:
	//   To validate against allOf, the given data
	//   must be valid against all of the given subschemas.
	//
	// Current implementation simply select the strictest constraints from both schemes.
	//
	// Note that this approach will not work with different 'pattern' or 'multipleOf' constraints
	// because they cannot be merged.
	switch s1.Type {
	case jsonschema.String:
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

		return r, nil
	case jsonschema.Integer, jsonschema.Number:
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
			return nil, errors.Errorf("multipleOf is different")
		}
		r.MultipleOf = s1.MultipleOf
		return r, nil

	case jsonschema.Array:
		r.Item, err = mergeSchemes(s1.Item, s2.Item)
		if err != nil {
			return nil, errors.Wrap(err, "merge item schema")
		}

		r.MinItems = someU64(s1.MinItems, s2.MinItems, selectMaxU64)
		r.MaxItems = someU64(s1.MaxItems, s2.MaxItems, selectMinU64)
		r.UniqueItems = s1.UniqueItems || s2.UniqueItems
		return r, nil

	case jsonschema.Null, jsonschema.Boolean:
		return r, nil

	case jsonschema.Object:
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
		case s1.AdditionalProperties != nil && s2.AdditionalProperties != nil:
			return nil, &ErrNotImplemented{Name: "allOf additionalProperties merging"}
		}

		r.MinProperties = someU64(s1.MinProperties, s2.MinProperties, selectMaxU64)
		r.MaxProperties = someU64(s1.MaxProperties, s2.MaxProperties, selectMinU64)
		r.Properties, err = mergeProperties(s1.Properties, s2.Properties)
		if err != nil {
			return nil, errors.Wrap(err, "merge properties")
		}

		return r, nil

	default:
		return nil, &ErrNotImplemented{Name: "complex schema merging"}
	}
}

// mergeProperties finds properties with identical names
// and tries to merge them into one, avoiding duplicates.
func mergeProperties(p1, p2 []jsonschema.Property) ([]jsonschema.Property, error) {
	var (
		propmap    = make(map[string]jsonschema.Property, len(p1)+len(p2))
		order      = make(map[string]int, len(p1)+len(p2))
		orderIndex = 0
	)

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
				Description: "Merged property",
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
		result[order[name]] = p
	}

	return result, nil
}

func mergeEnums(s1, s2 *jsonschema.Schema) ([]interface{}, error) {
	switch {
	case len(s1.Enum) == 0 && len(s2.Enum) == 0:
		return nil, nil
	case len(s1.Enum) > 0 && len(s2.Enum) == 0:
		return s1.Enum, nil
	case len(s1.Enum) == 0 && len(s2.Enum) > 0:
		return s2.Enum, nil
	}

	// TODO: Merge enums and check for duplicates.
	return nil, &ErrNotImplemented{Name: "allOf enum merging"}
}
