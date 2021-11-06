package gen

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/ogen-go/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

type schemaGen struct {
	side       []*ir.Type
	localRefs  map[string]*ir.Type
	globalRefs map[string]*ir.Type
}

func (g *schemaGen) generate(name string, schema *oas.Schema) (*ir.Type, error) {
	if ref := schema.Ref; ref != "" {
		if t, ok := g.globalRefs[ref]; ok {
			return t, nil
		}
		if t, ok := g.localRefs[ref]; ok {
			return t, nil
		}

		name = pascal(strings.TrimPrefix(ref, "#/components/schemas/"))
	}

	switch {
	case len(schema.AnyOf) > 0:
		return nil, &ErrNotImplemented{"anyOf"}
	case len(schema.AllOf) > 0:
		return nil, &ErrNotImplemented{"allOf"}
	}

	var reg *regexp.Regexp
	if schema.Pattern != "" {
		p, err := regexp.Compile(schema.Pattern)
		if err != nil {
			return nil, errors.Wrap(err, "pattern")
		}
		reg = p
	}

	side := func(t *ir.Type) *ir.Type {
		// Set validation fields.
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

		if schema.MaxItems != nil {
			t.Validators.Array.SetMaxLength(int(*schema.MaxItems))
		}
		if schema.MinItems != nil {
			t.Validators.Array.SetMinLength(int(*schema.MinItems))
		}

		t.Validators.String.Regex = reg
		if schema.MaxLength != nil {
			t.Validators.String.SetMaxLength(int(*schema.MaxLength))
		}
		if schema.MinLength != nil {
			t.Validators.String.SetMinLength(int(*schema.MinLength))
		}
		if t.Schema.Format == oas.FormatEmail {
			t.Validators.String.Email = true
		}
		if t.Schema.Format == oas.FormatHostname {
			t.Validators.String.Hostname = true
		}
		if ref := t.Schema.Ref; ref != "" {
			if t.Is(ir.KindPrimitive, ir.KindArray) {
				t = ir.Alias(name, t)
			}

			g.localRefs[ref] = t
			return t
		}

		if t.Is(ir.KindStruct, ir.KindEnum, ir.KindSum) {
			g.side = append(g.side, t)
		}

		return t
	}

	switch schema.Type {
	case oas.Object:
		s := side(&ir.Type{
			Kind:   ir.KindStruct,
			Name:   name,
			Schema: schema,
		})

		for i := range schema.Properties {
			prop := schema.Properties[i]
			typ, err := g.generate(pascalMP(name, prop.Name), prop.Schema)
			if err != nil {
				return nil, errors.Wrapf(err, "field %s", prop.Name)
			}

			s.Fields = append(s.Fields, &ir.Field{
				Name: pascalMP(prop.Name),
				Type: typ,
				Tag: ir.Tag{
					JSON: prop.Name,
				},
				Spec: &prop,
			})
		}

		return s, nil

	case oas.Array:
		array := &ir.Type{
			Kind:        ir.KindArray,
			Schema:      schema,
			NilSemantic: ir.NilInvalid,
		}

		ret := side(array)
		item, err := g.generate(name+"Item", schema.Item)
		if err != nil {
			return nil, err
		}

		array.Item = item
		return ret, nil

	case oas.String, oas.Integer, oas.Number, oas.Boolean:
		prim, err := g.primitive(name, schema)
		if err != nil {
			return nil, err
		}

		return side(prim), nil

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
			var result []rune
			for i, c := range t.Go() {
				if i == 0 {
					c = unicode.ToUpper(c)
				}
				result = append(result, c)
			}
			t.Name = string(result)
			if _, ok := names[t.Name]; ok {
				return nil, errors.Wrap(&ErrNotImplemented{
					Name: "sum types with same names",
				}, name)
			}
			names[t.Name] = struct{}{}
			sum.SumOf = append(sum.SumOf, t)
		}
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

		// Determine unique fields for each SumOf variant.
		uniq := map[string]map[string]struct{}{}
		var isComplex bool
		for _, s := range sum.SumOf {
			uniq[s.Name] = map[string]struct{}{}
			if !s.Is(ir.KindPrimitive) {
				isComplex = true
			}
			for _, f := range s.Fields {
				uniq[s.Name][f.Name] = struct{}{}
			}
		}
		for _, s := range sum.SumOf {
			k := s.Name
			f := uniq[k]
			for _, otherS := range sum.SumOf {
				otherK := otherS.Name
				if otherK == k {
					continue
				}
				for otherField := range uniq[otherK] {
					delete(f, otherField)
				}
			}
			uniq[k] = f
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
			if isComplex && len(v.Unique) == 0 {
				// Unable to deterministically select sub-schema only on fields.
				return nil, errors.Wrapf(&ErrNotImplemented{Name: "discriminator inference"},
					"oneOf %s: variant %s: no unique fields, "+
						"unable to parse without discriminator", sum.Name, k,
				)
			}
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
	case oas.Stream:
		return ir.Embedded(), nil
	default:
		panic("unreachable")
	}
}

func (g *schemaGen) primitive(name string, schema *oas.Schema) (*ir.Type, error) {
	typ, err := parseSimple(schema)
	if err != nil {
		return nil, err
	}

	if len(schema.Enum) > 0 {
		if !typ.Is(ir.KindPrimitive) {
			return nil, errors.Errorf("unsupported enum type: %q", schema.Type)
		}

		var variants []*ir.EnumVariant
		for _, v := range schema.Enum {
			vstr := fmt.Sprintf("%v", v)
			if vstr == "" {
				vstr = "Empty"
			}

			variants = append(variants, &ir.EnumVariant{
				Name:  pascalMP(name, vstr),
				Value: v,
			})
		}

		return &ir.Type{
			Kind:         ir.KindEnum,
			Name:         name,
			Primitive:    typ.Primitive,
			EnumVariants: variants,
			Schema:       schema,
		}, nil
	}

	return typ, nil
}

func parseSimple(schema *oas.Schema) (*ir.Type, error) {
	typ, format := schema.Type, schema.Format
	switch typ {
	case oas.Integer:
		switch format {
		case oas.FormatInt32:
			return ir.Primitive(ir.Int32, schema), nil
		case oas.FormatInt64:
			return ir.Primitive(ir.Int64, schema), nil
		case oas.FormatNone:
			return ir.Primitive(ir.Int, schema), nil
		default:
			return nil, errors.Errorf("unexpected integer format: %q", format)
		}
	case oas.Number:
		switch format {
		case oas.FormatFloat:
			return ir.Primitive(ir.Float32, schema), nil
		case oas.FormatDouble, oas.FormatNone:
			return ir.Primitive(ir.Float64, schema), nil
		default:
			return nil, errors.Errorf("unexpected number format: %q", format)
		}
	case oas.String:
		switch format {
		case oas.FormatByte:
			return ir.Array(ir.Primitive(ir.Byte, nil), ir.NilInvalid, schema), nil
		case oas.FormatDateTime, oas.FormatDate, oas.FormatTime:
			return ir.Primitive(ir.Time, schema), nil
		case oas.FormatDuration:
			return ir.Primitive(ir.Duration, schema), nil
		case oas.FormatUUID:
			return ir.Primitive(ir.UUID, schema), nil
		case oas.FormatIP, oas.FormatIPv4, oas.FormatIPv6:
			return ir.Primitive(ir.IP, schema), nil
		case oas.FormatURI:
			return ir.Primitive(ir.URL, schema), nil
		case oas.FormatPassword, oas.FormatNone:
			return ir.Primitive(ir.String, schema), nil
		default:
			// return nil, errors.Errorf("unexpected string format: %q", format)
			return ir.Primitive(ir.String, schema), nil
		}
	case oas.Boolean:
		switch format {
		case oas.FormatNone:
			return ir.Primitive(ir.Bool, schema), nil
		default:
			return nil, errors.Errorf("unexpected bool format: %q", format)
		}
	default:
		return nil, errors.Errorf("unexpected type: %q", typ)
	}
}
