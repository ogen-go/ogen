package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/capitalize"
	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

type schemaGen struct {
	side      []*ir.Type
	localRefs map[string]*ir.Type
	lookupRef func(ref string) (*ir.Type, bool)
	nameRef   func(ref string) (string, error)
}

func newSchemaGen(lookupRef func(ref string) (*ir.Type, bool)) *schemaGen {
	return &schemaGen{
		side:      nil,
		localRefs: map[string]*ir.Type{},
		lookupRef: lookupRef,
		nameRef: func(ref string) (string, error) {
			name, err := pascal(strings.TrimPrefix(ref, "#/components/schemas/"))
			if err != nil {
				return "", err
			}
			return name, nil
		},
	}
}

func variantFieldName(t *ir.Type) string {
	return capitalize.Capitalize(t.NamePostfix())
}

func (g *schemaGen) generate(name string, schema *jsonschema.Schema) (_ *ir.Type, err error) {
	if schema == nil {
		return nil, &ErrNotImplemented{Name: "empty schema"}
	}

	if ref := schema.Ref; ref != "" {
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
	if schema.DefaultSet && schema.Type == jsonschema.Object {
		return nil, &ErrNotImplemented{Name: "object defaults"}
	}

	if name[0] >= '0' && name[0] <= '9' {
		name = "R" + name
	}

	side := func(t *ir.Type) *ir.Type {
		if t.Schema != nil {
			if ref := t.Schema.Ref; ref != "" {
				if t.Is(ir.KindPrimitive, ir.KindArray) {
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

	switch {
	case len(schema.AnyOf) > 0:
		t, err := g.anyOf(name, schema)
		if err != nil {
			return nil, errors.Wrap(err, "anyOf")
		}
		return side(t), nil
	case len(schema.AllOf) > 0:
		return nil, &ErrNotImplemented{"allOf"}
	case len(schema.OneOf) > 0:
		t, err := g.oneOf(name, schema)
		if err != nil {
			return nil, errors.Wrap(err, "oneOf")
		}
		return side(t), nil
	}

	switch schema.Type {
	case jsonschema.Object:
		kind := ir.KindStruct
		if schema.AdditionalProperties {
			kind = ir.KindMap
		}

		s := side(&ir.Type{
			Kind:   kind,
			Name:   name,
			Schema: schema,
		})
		s.Validators.SetObject(schema)

		for i := range schema.Properties {
			prop := schema.Properties[i]
			propTypeName, err := pascalSpecial(name, prop.Name)
			if err != nil {
				return nil, errors.Wrapf(err, "property type name: %q", prop.Name)
			}

			t, err := g.generate(propTypeName, prop.Schema)
			if err != nil {
				return nil, errors.Wrapf(err, "field %s", prop.Name)
			}

			propertyName := strings.TrimSpace(prop.Name)
			if propertyName == "" {
				propertyName = fmt.Sprintf("Field%d", i)
			}

			fieldName, err := pascalSpecial(propertyName)
			if err != nil {
				return nil, errors.Wrapf(err, "property name: %q", propertyName)
			}

			s.Fields = append(s.Fields, &ir.Field{
				Name: fieldName,
				Type: t,
				Tag: ir.Tag{
					JSON: prop.Name,
				},
				Spec: &prop,
			})
		}

		if schema.AdditionalProperties {
			if schema.Item != nil {
				s.Item, err = g.generate(name+"Item", schema.Item)
				if err != nil {
					return nil, errors.Wrap(err, "item")
				}
			} else {
				s.Item = ir.Any()
			}
		}

		return s, nil
	case jsonschema.Array:
		array := &ir.Type{
			Kind:        ir.KindArray,
			Schema:      schema,
			NilSemantic: ir.NilInvalid,
		}

		array.Validators.SetArray(schema)

		ret := side(array)
		if schema.Item != nil {
			array.Item, err = g.generate(name+"Item", schema.Item)
			if err != nil {
				return nil, errors.Wrap(err, "item")
			}
		} else {
			array.Item = ir.Any()
		}

		return ret, nil

	case jsonschema.String, jsonschema.Integer, jsonschema.Number, jsonschema.Boolean:
		t, err := g.primitive(name, schema)
		if err != nil {
			return nil, errors.Wrap(err, "primitive")
		}

		switch schema.Type {
		case jsonschema.String:
			if err := t.Validators.SetString(schema); err != nil {
				return nil, errors.Wrap(err, "string validator")
			}
		case jsonschema.Integer:
			if err := t.Validators.SetInt(schema); err != nil {
				return nil, errors.Wrap(err, "int validator")
			}
		case jsonschema.Number:
			if err := t.Validators.SetFloat(schema); err != nil {
				return nil, errors.Wrap(err, "float validator")
			}
		}

		return side(t), nil
	case jsonschema.Empty:
		return side(ir.Any()), nil
	default:
		panic(unreachable(schema.Type))
	}
}
