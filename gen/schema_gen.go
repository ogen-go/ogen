package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/internal/naming"
	"github.com/ogen-go/ogen/jsonschema"
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

	customFormats map[jsonschema.SchemaType]map[string]ir.CustomFormat
	depthLimit    int
	depthCount    int

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
		depthLimit:    defaultSchemaDepthLimit,
		customFormats: map[jsonschema.SchemaType]map[string]ir.CustomFormat{},
		log:           zap.NewNop(),
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
					nameDefinedAt: prop.X.Pointer.Field("name"),
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
					nameDefinedAt: schema.Pointer.Key(key),
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
			if t.Validators.Int.Set() {
				switch t.Primitive {
				case ir.Int,
					ir.Int8,
					ir.Int16,
					ir.Int32,
					ir.Int64,
					ir.Uint,
					ir.Uint8,
					ir.Uint16,
					ir.Uint32,
					ir.Uint64:
				default:
					g.log.Warn("Int validator cannot be applied to generated type and will be ignored", fields...)
				}
			}
		case jsonschema.Number:
			if err := t.Validators.SetFloat(schema); err != nil {
				return nil, errors.Wrap(err, "float validator")
			}
			if t.Validators.Float.Set() {
				switch t.Primitive {
				case ir.Float32, ir.Float64:
				default:
					g.log.Warn("Float validator cannot be applied to generated type and will be ignored", fields...)
				}
			}
		}

		return g.regtype(name, t), nil
	case jsonschema.Empty:
		if format, ok := g.customFormats[schema.Type][schema.Format]; ok {
			return g.customFormat(format, schema), nil
		}

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
