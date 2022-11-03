package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/naming"
	"github.com/ogen-go/ogen/jsonschema"
)

type schemaGen struct {
	side      []*ir.Type
	localRefs map[string]*ir.Type
	lookupRef func(ref string) (*ir.Type, bool)
	nameRef   func(ref string) (string, error)
	fail      func(err error) error

	log *zap.Logger
}

func newSchemaGen(lookupRef func(ref string) (*ir.Type, bool)) *schemaGen {
	return &schemaGen{
		side:      nil,
		localRefs: map[string]*ir.Type{},
		lookupRef: lookupRef,
		nameRef: func(ref string) (string, error) {
			name, err := pascal(cleanRef(ref))
			if err != nil {
				return "", err
			}
			return name, nil
		},
		fail: func(err error) error {
			return err
		},
		log: zap.NewNop(),
	}
}

func variantFieldName(t *ir.Type) string {
	return naming.Capitalize(t.NamePostfix())
}

func (g *schemaGen) generate(name string, schema *jsonschema.Schema, optional bool) (*ir.Type, error) {
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

	if schema.DefaultSet {
		var implErr error
		switch schema.Type {
		case jsonschema.Object:
			implErr = &ErrNotImplemented{Name: "object defaults"}
		case jsonschema.Array:
			implErr = &ErrNotImplemented{Name: "array defaults"}
		}
		// Do not fail schema generation if we cannot handle defaults.
		if err := g.fail(implErr); err != nil {
			return nil, err
		}
		schema.DefaultSet = implErr == nil
	}

	if n := schema.XOgenName; n != "" {
		name = n
	} else if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "R" + name
	}

	switch {
	case len(schema.AnyOf) > 0:
		t, err := g.anyOf(name, schema)
		if err != nil {
			return nil, errors.Wrap(err, "anyOf")
		}
		return t, nil
	case len(schema.AllOf) > 0:
		t, err := g.allOf(name, schema)
		if err != nil {
			return nil, errors.Wrap(err, "allOf")
		}
		return t, nil
	case len(schema.OneOf) > 0:
		t, err := g.oneOf(name, schema)
		if err != nil {
			return nil, errors.Wrap(err, "oneOf")
		}
		return t, nil
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
				// TODO(tdakkota): check name for collision.
				s.Fields = append(s.Fields, &ir.Field{
					Name:   "AdditionalProps",
					Type:   mapType,
					Inline: ir.InlineAdditional,
				})
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
					// TODO(tdakkota): check name for collision.
					s.Fields = append(s.Fields, &ir.Field{
						Name:   suffix + "Props",
						Type:   mapType,
						Inline: ir.InlinePattern,
					})
				}
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

		ret := g.regtype(name, array)
		if schema.Item != nil {
			array.Item, err = g.generate(name+"Item", schema.Item, false)
			if err != nil {
				return nil, errors.Wrap(err, "item")
			}
		} else {
			array.Item = ir.Any(schema.Item)
		}

		return ret, nil

	case jsonschema.String, jsonschema.Integer, jsonschema.Number, jsonschema.Boolean, jsonschema.Null:
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
		if ref := t.Schema.Ref; ref != "" {
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
