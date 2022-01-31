package gen

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

type schemaGen struct {
	side       []*ir.Type
	localRefs  map[string]*ir.Type
	globalRefs map[string]*ir.Type
}

func variantFieldName(t *ir.Type) string {
	capitalize := func(s string) string {
		r, size := utf8.DecodeRuneInString(s)
		return string(unicode.ToUpper(r)) + s[size:]
	}

	var result string
	switch t.Kind {
	case ir.KindArray:
		result = "Array" + variantFieldName(t.Item)
	case ir.KindPointer:
		result = variantFieldName(t.PointerTo)
	default:
		result = t.Go()
	}
	return capitalize(result)
}

func (g *schemaGen) generate(name string, schema *oas.Schema) (_ *ir.Type, err error) {
	if ref := schema.Ref; ref != "" {
		if t, ok := g.globalRefs[ref]; ok {
			return t, nil
		}
		if t, ok := g.localRefs[ref]; ok {
			return t, nil
		}

		name = pascal(strings.TrimPrefix(ref, "#/components/schemas/"))
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "R" + name
	}

	switch {
	case len(schema.AnyOf) > 0:
		return nil, &ErrNotImplemented{"anyOf"}
	case len(schema.AllOf) > 0:
		return nil, &ErrNotImplemented{"allOf"}
	}

	side := func(t *ir.Type) *ir.Type {
		if ref := t.Schema.Ref; ref != "" {
			if t.Is(ir.KindPrimitive, ir.KindArray) {
				t = ir.Alias(name, t)
			}

			g.localRefs[ref] = t
			return t
		}

		if t.Is(ir.KindStruct, ir.KindMap, ir.KindEnum, ir.KindSum) {
			g.side = append(g.side, t)
		}

		return t
	}

	switch schema.Type {
	case oas.Object:
		kind := ir.KindStruct
		if schema.Item != nil {
			kind = ir.KindMap
		}

		s := side(&ir.Type{
			Kind:   kind,
			Name:   name,
			Schema: schema,
		})

		for i := range schema.Properties {
			prop := schema.Properties[i]
			t, err := g.generate(pascalSpecial(name, prop.Name), prop.Schema)
			if err != nil {
				return nil, errors.Wrapf(err, "field %s", prop.Name)
			}

			s.Fields = append(s.Fields, &ir.Field{
				Name: pascalSpecial(prop.Name),
				Type: t,
				Tag: ir.Tag{
					JSON: prop.Name,
				},
				Spec: &prop,
			})
		}

		if schema.Item != nil {
			s.Item, err = g.generate(name+"Item", schema.Item)
			if err != nil {
				return nil, err
			}
		}

		return s, nil
	case oas.Array:
		array := &ir.Type{
			Kind:        ir.KindArray,
			Schema:      schema,
			NilSemantic: ir.NilInvalid,
		}

		if schema.MaxItems != nil {
			array.Validators.Array.SetMaxLength(int(*schema.MaxItems))
		}
		if schema.MinItems != nil {
			array.Validators.Array.SetMinLength(int(*schema.MinItems))
		}

		ret := side(array)
		array.Item, err = g.generate(name+"Item", schema.Item)
		if err != nil {
			return nil, err
		}

		return ret, nil

	case oas.String, oas.Integer, oas.Number, oas.Boolean:
		t, err := g.primitive(name, schema)
		if err != nil {
			return nil, err
		}

		switch schema.Type {
		case oas.String:
			if schema.Pattern != "" {
				t.Validators.String.Regex, err = regexp.Compile(schema.Pattern)
				if err != nil {
					return nil, errors.Wrap(err, "pattern")
				}
			}
			if schema.MaxLength != nil {
				t.Validators.String.SetMaxLength(int(*schema.MaxLength))
			}
			if schema.MinLength != nil {
				t.Validators.String.SetMinLength(int(*schema.MinLength))
			}
			if schema.Format == oas.FormatEmail {
				t.Validators.String.Email = true
			}
			if schema.Format == oas.FormatHostname {
				t.Validators.String.Hostname = true
			}

		case oas.Integer, oas.Number:
			if schema.MultipleOf != nil {
				t.Validators.Int.MultipleOf = *schema.MultipleOf
				t.Validators.Int.MultipleOfSet = true
			}
			if schema.Maximum != nil {
				t.Validators.Int.Max = *schema.Maximum
				t.Validators.Int.MaxSet = true
			}
			if schema.Minimum != nil {
				t.Validators.Int.Min = *schema.Minimum
				t.Validators.Int.MinSet = true
			}
			t.Validators.Int.MaxExclusive = schema.ExclusiveMaximum
			t.Validators.Int.MinExclusive = schema.ExclusiveMinimum
		}

		return side(t), nil

	case oas.Empty:
		sum := &ir.Type{
			Name:   name,
			Kind:   ir.KindSum,
			Schema: schema,
		}
		names := map[string]struct{}{}
		for i, s := range schema.OneOf {
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
			sum.SumOf = append(sum.SumOf, t)
		}

		// 1st case: explicit discriminator.
		if d := schema.Discriminator; d != nil {
			sum.SumSpec.Discriminator = schema.Discriminator.PropertyName
			for k, v := range schema.Discriminator.Mapping {
				// Explicit mapping.
				var found bool
				for _, s := range sum.SumOf {
					if path.Base(s.Schema.Ref) == v {
						found = true
						sum.SumSpec.Mapping = append(sum.SumSpec.Mapping, ir.SumSpecMap{
							Key:  k,
							Type: s.Name,
						})
					}
				}
				if !found {
					return nil, errors.Errorf("discriminator: unable to map %s to %s", k, v)
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
			return side(sum), nil
		}

		// 2nd case: distinguish by serialization type.
		var (
			// Collect map of variant kinds.
			typeMap = map[ir.TypeDiscriminator]struct{}{}
			// If all variants have different kinds, so
			// we can distinguish them by JSON type.
			canUseTypeDiscriminator = true
		)
		for _, s := range sum.SumOf {
			var kind ir.TypeDiscriminator
			kind.Set(s)
			if _, ok := typeMap[kind]; ok {
				// Type kind is not unique, so we can distinguish variants by type.
				canUseTypeDiscriminator = false
				break
			}
			typeMap[kind] = struct{}{}
		}
		if canUseTypeDiscriminator {
			sum.SumSpec.TypeDiscriminator = true
			return side(sum), nil
		}

		// 3rd case: distinguish by unique fields.
		var (
			// Determine unique fields for each SumOf variant.
			uniq = map[string]map[string]struct{}{}
			// Whether sum has complex types.
			isComplex bool
		)
		for _, s := range sum.SumOf {
			uniq[s.Name] = map[string]struct{}{}
			if !s.Is(ir.KindPrimitive) {
				isComplex = true
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

			if isComplex {
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
		if !isComplex {
			return side(sum), nil
		}
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
		return side(sum), nil
	default:
		panic("unreachable")
	}
}

func (g *schemaGen) primitive(name string, schema *oas.Schema) (*ir.Type, error) {
	t, err := parseSimple(schema)
	if err != nil {
		return nil, err
	}

	if len(schema.Enum) > 0 {
		if !t.Is(ir.KindPrimitive) {
			return nil, errors.Errorf("unsupported enum type: %q", schema.Type)
		}

		hasDuplicateNames := func() bool {
			names := map[string]struct{}{}
			for _, v := range schema.Enum {
				vstr := fmt.Sprintf("%v", v)
				if vstr == "" {
					vstr = "Empty"
				}

				k := pascalSpecial(name, vstr)
				if _, ok := names[k]; ok {
					return true
				}
				names[k] = struct{}{}
			}

			return false
		}()

		var variants []*ir.EnumVariant
		for _, v := range schema.Enum {
			vstr := fmt.Sprintf("%v", v)
			if vstr == "" {
				vstr = "Empty"
			}

			var variantName string
			if hasDuplicateNames {
				variantName = name + "_" + vstr
			} else {
				variantName = pascalSpecial(name, vstr)
			}

			variants = append(variants, &ir.EnumVariant{
				Name:  variantName,
				Value: v,
			})
		}

		return &ir.Type{
			Kind:         ir.KindEnum,
			Name:         name,
			Primitive:    t.Primitive,
			EnumVariants: variants,
			Schema:       schema,
		}, nil
	}

	return t, nil
}

func parseSimple(schema *oas.Schema) (*ir.Type, error) {
	mapping := map[oas.SchemaType]map[oas.Format]ir.PrimitiveType{
		oas.Integer: {
			oas.FormatInt32: ir.Int32,
			oas.FormatInt64: ir.Int64,
			oas.FormatNone:  ir.Int,
		},
		oas.Number: {
			oas.FormatFloat:  ir.Float32,
			oas.FormatDouble: ir.Float64,
			oas.FormatNone:   ir.Float64,
			oas.FormatInt32:  ir.Int32,
			oas.FormatInt64:  ir.Int64,
		},
		oas.String: {
			oas.FormatByte:     ir.ByteSlice,
			oas.FormatDateTime: ir.Time,
			oas.FormatDate:     ir.Time,
			oas.FormatTime:     ir.Time,
			oas.FormatDuration: ir.Duration,
			oas.FormatUUID:     ir.UUID,
			oas.FormatIP:       ir.IP,
			oas.FormatIPv4:     ir.IP,
			oas.FormatIPv6:     ir.IP,
			oas.FormatURI:      ir.URL,
			oas.FormatPassword: ir.String,
			oas.FormatNone:     ir.String,
		},
		oas.Boolean: {
			oas.FormatNone: ir.Bool,
		},
	}

	t, found := mapping[schema.Type][schema.Format]
	if !found {
		// Return string type for unknown string formats.
		if schema.Type == oas.String {
			return ir.Primitive(ir.String, schema), nil
		}
		return nil, errors.Errorf("unexpected %q format: %q", schema.Type, schema.Format)
	}

	return ir.Primitive(t, schema), nil
}
