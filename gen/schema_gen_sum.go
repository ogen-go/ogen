package gen

import (
	"bytes"
	"fmt"
	"path"
	"sort"

	"github.com/go-faster/errors"

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

func (g *schemaGen) collectSumVariants(
	name string,
	parent *jsonschema.Schema,
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
		if ref := s.Ref; ref != "" && ref == parent.Ref {
			return nil, errors.Errorf("reference %q [%d] leads to infinite recursion", ref, i)
		}

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
	sum := g.regtype(name, &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: schema,
	})
	{
		variants, err := g.collectSumVariants(name, schema, schema.AnyOf)
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
	sum := g.regtype(name, &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: schema,
	})
	{
		variants, err := g.collectSumVariants(name, schema, schema.OneOf)
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
	mergedSchema, err := mergeNSchemes(schema.AllOf)
	if err != nil {
		return nil, err
	}

	mergedSchema.Ref = schema.Ref
	return g.generate2(name, mergedSchema)
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

	nullable := s1.Nullable
	if s2.Nullable {
		nullable = s2.Nullable
	}

	r := &jsonschema.Schema{
		Type:        s1.Type,
		Format:      s1.Format,
		Nullable:    nullable,
		Description: "Merged schema",
	}

	switch s1.Type {
	case jsonschema.String:
		if s1.MaxLength != s2.MaxLength {
			return nil, errors.Errorf("maxLength is different")
		}
		r.MaxLength = s1.MaxLength

		if s1.MinLength != s2.MinLength {
			return nil, errors.Errorf("minLength is different")
		}
		r.MinLength = s1.MinLength

		if s1.Pattern != s2.Pattern {
			return nil, errors.Errorf("pattern is different")
		}
		r.Pattern = s1.Pattern

		return r, nil
	case jsonschema.Integer, jsonschema.Number:
		if !bytes.Equal(s1.Maximum, s2.Maximum) {
			return nil, errors.Errorf("maximum is different")
		}
		r.Maximum = s1.Maximum

		if s1.ExclusiveMaximum != s2.ExclusiveMaximum {
			return nil, errors.Errorf("exclusiveMaximum is different")
		}
		r.ExclusiveMaximum = s1.ExclusiveMaximum

		if !bytes.Equal(s1.Minimum, s2.Minimum) {
			return nil, errors.Errorf("minimum is different")
		}
		r.Minimum = s1.Minimum

		if s1.ExclusiveMinimum != s2.ExclusiveMinimum {
			return nil, errors.Errorf("exclusiveMinimum is different")
		}
		r.ExclusiveMinimum = s1.ExclusiveMinimum

		if !bytes.Equal(s1.MultipleOf, s2.MultipleOf) {
			return nil, errors.Errorf("multipleOf is different")
		}
		r.MultipleOf = s1.MultipleOf
		return r, nil

	case jsonschema.Array:
		r.Item, err = mergeSchemes(s1.Item, s2.Item)
		if err != nil {
			return nil, errors.Wrap(err, "item")
		}

		if s1.MinItems != s2.MinItems {
			return nil, errors.Wrap(err, "minItems is different")
		}
		r.MinItems = s1.MinItems

		if s1.MaxItems != s2.MaxItems {
			return nil, errors.Wrap(err, "maxItems is different")
		}
		r.MaxItems = s1.MaxItems

		if s1.UniqueItems != s2.UniqueItems {
			return nil, errors.Errorf("uniqueItems is different")
		}
		r.UniqueItems = s1.UniqueItems

		return r, nil

	case jsonschema.Null, jsonschema.Boolean:
		return r, nil

	case jsonschema.Object:
		r.Properties, err = mergeProperties([]*jsonschema.Schema{s1, s2})
		if err != nil {
			return nil, errors.Wrap(err, "merge properties")
		}

		return r, nil

	default:
		return nil, &ErrNotImplemented{Name: "complex schema merging"}
	}
}

func mergeProperties(schemas []*jsonschema.Schema) ([]jsonschema.Property, error) {
	propmap := make(map[string]jsonschema.Property)
	order := make(map[string]int)
	orderIndex := 0
	for _, s := range schemas {
		if s.Type != jsonschema.Object {
			return nil, &ErrNotImplemented{Name: "non-object schema type"}
		}
		for _, p := range s.Properties {
			if confP, ok := propmap[p.Name]; ok {
				// Property name conflict.
				s, err := mergeSchemes(p.Schema, confP.Schema)
				if err != nil {
					return nil, errors.Wrap(err, "try to merge conflicting property schemas")
				}

				required := p.Required
				if confP.Required {
					required = true
				}

				propmap[p.Name] = jsonschema.Property{
					Name:        p.Name,
					Description: "Merged property",
					Schema:      s,
					Required:    required,
				}
				continue
			}

			propmap[p.Name] = p
			order[p.Name] = orderIndex
			orderIndex++
		}
	}

	result := make([]jsonschema.Property, len(propmap))
	for name, p := range propmap {
		result[order[name]] = p
	}
	return result, nil
}
