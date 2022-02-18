package gen

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func canUseTypeDiscriminator(sum []*ir.Type) bool {
	// Collect map of variant kinds.
	typeMap := map[ir.TypeDiscriminator]struct{}{}

	for _, s := range sum {
		if s.IsAny() {
			// Cannot make typed sum with Any.
			return false
		}

		var kind ir.TypeDiscriminator
		kind.Set(s)
		if _, ok := typeMap[kind]; ok {
			// Type kind is not unique, so we can distinguish variants by type.
			return false
		}
		typeMap[kind] = struct{}{}
	}
	return true
}

func (g *schemaGen) collectSumVariants(name string, schemas []*jsonschema.Schema) (sum []*ir.Type, _ error) {
	names := map[string]struct{}{}
	for i, s := range schemas {
		t, err := g.generate(fmt.Sprintf("%s%d", name, i), s)
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
	sum := &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: schema,
	}
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
	sum := &ir.Type{
		Name:   name,
		Kind:   ir.KindSum,
		Schema: schema,
	}
	{
		variants, err := g.collectSumVariants(name, schema.OneOf)
		if err != nil {
			return nil, errors.Wrap(err, "collect variants")
		}
		sum.SumOf = variants
	}

	// 1st case: explicit discriminator.
	if d := schema.Discriminator; d != nil {
		sum.SumSpec.Discriminator = schema.Discriminator.PropertyName
		for k, v := range schema.Discriminator.Mapping {
			// Explicit mapping.
			var found bool
			for _, s := range sum.SumOf {
				if s.Schema.Ref == v || path.Base(s.Schema.Ref) == v {
					found = true
					sum.SumSpec.Mapping = append(sum.SumSpec.Mapping, ir.SumSpecMap{
						Key:  k,
						Type: s.Name,
					})
				}
			}
			if !found {
				return nil, errors.Errorf("discriminator: unable to map %q to %q", k, v)
			}
		}
		if len(sum.SumSpec.Mapping) == 0 {
			// Implicit mapping, defaults to type name.
			for _, s := range sum.SumOf {
				sum.SumSpec.Mapping = append(sum.SumSpec.Mapping, ir.SumSpecMap{
					Key:  path.Base(s.Schema.Ref),
					Type: s.Name,
				})
			}
		}
		sort.SliceStable(sum.SumSpec.Mapping, func(i, j int) bool {
			a := sum.SumSpec.Mapping[i]
			b := sum.SumSpec.Mapping[j]
			return strings.Compare(a.Key, b.Key) < 0
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
		if !s.Is(ir.KindMap, ir.KindStruct) {
			return nil, errors.Wrapf(&ErrNotImplemented{Name: "discriminator inference"},
				"oneOf %s: variant %s: no unique fields, "+
					"unable to parse without discriminator", sum.Name, s.Name,
			)
		}
		for _, f := range s.Fields {
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
		return strings.Compare(a.Name, b.Name) < 0
	})
	for _, v := range variants {
		for _, s := range sum.SumOf {
			if s.Name != v.Name {
				continue
			}
			if len(s.SumSpec.Unique) > 0 {
				continue
			}
			for _, f := range s.Fields {
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
