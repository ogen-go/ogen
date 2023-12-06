// Package jsonschema contains parser for JSON Schema.
package jsonschema

import (
	"encoding/json"
	"fmt"
	"go/token"
	"math/big"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/exp/slices"

	ogenjson "github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/ogenregex"
)

const (
	xOgenName       = "x-ogen-name"
	xOgenProperties = "x-ogen-properties"
	xOapiExtraTags  = "x-oapi-codegen-extra-tags"
)

// Parser parses JSON schemas.
type Parser struct {
	external ExternalResolver
	schemas  map[string]resolver
	refcache map[jsonpointer.RefKey]*Schema

	rootFile location.File // optional, used for error messages

	inferTypes bool
}

// NewParser creates new Parser.
func NewParser(s Settings) *Parser {
	s.setDefaults()
	return &Parser{
		external: s.External,
		schemas: map[string]resolver{
			"": {
				ReferenceResolver: s.Resolver,
				file:              s.File,
			},
		},
		refcache:   map[jsonpointer.RefKey]*Schema{},
		rootFile:   s.File,
		inferTypes: s.InferTypes,
	}
}

// Parse parses given RawSchema and returns parsed Schema.
func (p *Parser) Parse(schema *RawSchema, ctx *jsonpointer.ResolveCtx) (*Schema, error) {
	return p.parse(schema, ctx)
}

// Resolve resolves Schema by given ref.
func (p *Parser) Resolve(ref string, ctx *jsonpointer.ResolveCtx) (*Schema, error) {
	return p.resolve(ref, ctx)
}

func (p *Parser) parse(schema *RawSchema, ctx *jsonpointer.ResolveCtx) (_ *Schema, rerr error) {
	if schema != nil {
		defer func() {
			rerr = p.wrapLocation(p.file(ctx), schema.Common.Locator, rerr)
		}()
	}
	return p.parse1(schema, ctx, func(s *Schema) *Schema {
		return p.extendInfo(schema, s, p.file(ctx))
	})
}

func (p *Parser) parse1(schema *RawSchema, ctx *jsonpointer.ResolveCtx, hook func(*Schema) *Schema) (*Schema, error) {
	s, err := p.parseSchema(schema, ctx, hook)
	if err != nil {
		return nil, err
	}

	if schema == nil || s == nil {
		return s, nil
	}

	if rd := schema.Discriminator; rd != nil {
		d, err := p.parseDiscriminator(rd, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "parse discriminator")
		}
		s.Discriminator = d
	}

	if enum := schema.Enum; len(enum) > 0 {
		loc := schema.Common.Locator.Field("enum")
		for i, a := range enum {
			for j, b := range enum {
				if i == j {
					continue
				}
				if ok, _ := ogenjson.Equal(a, b); ok {
					me := new(location.MultiError)
					me.Report(p.file(ctx), loc.Index(i), fmt.Sprintf("duplicate enum value: %q", a))
					me.Report(p.file(ctx), loc.Index(j), "")
					return nil, me
				}
			}
		}

		values, err := parseEnumValues(s, enum)
		if err != nil {
			err := errors.Wrap(err, "parse enum values")
			return nil, p.wrapLocation(p.file(ctx), loc, err)
		}
		s.Enum = values
		handleNullableEnum(s)
	}
	if d := schema.Default; len(d) > 0 {
		if err := func() error {
			v, err := parseJSONValue(nil, json.RawMessage(d))
			if err != nil {
				return err
			}

			s.Default = v
			s.DefaultSet = true
			return nil
		}(); err != nil {
			err := errors.Wrap(err, "parse default")
			return nil, p.wrapField("default", p.file(ctx), schema.Common.Locator, err)
		}
	}

	for key, val := range schema.Common.Extensions {
		if err := func() error {
			locator := schema.Common.Locator.Field(key)

			switch key {
			case xOgenName:
				if err := val.Decode(&s.XOgenName); err != nil {
					return err
				}

				if err := validateGoIdent(s.XOgenName); err != nil {
					return p.wrapLocation(p.file(ctx), locator, err)
				}
			case xOgenProperties:
				props := map[string]XProperty{}
				if err := val.Decode(&props); err != nil {
					return err
				}

				fieldNames := map[string]location.Pointer{}
				for propName, x := range props {
					// FIXME(tdakkota): linear search
					idx := slices.IndexFunc(s.Properties, func(p Property) bool { return p.Name == propName })
					if idx < 0 {
						err := errors.Errorf("unknown property %q", propName)
						return p.wrapLocation(p.file(ctx), locator.Key(propName), err)
					}

					if n := x.Name; n != nil {
						locator := locator.Field(propName).Field("name")
						if err := validateGoIdent(*n); err != nil {
							return p.wrapLocation(p.file(ctx), locator, err)
						}

						ptr := locator.Pointer(p.file(ctx))
						if existing, ok := fieldNames[*n]; ok {
							me := new(location.MultiError)
							me.ReportPtr(existing, fmt.Sprintf("duplicate field name %q", *n))
							me.ReportPtr(ptr, "")
							return me
						}
						fieldNames[*n] = ptr
					}

					x.Pointer = locator.Field(propName).Pointer(p.file(ctx))
					s.Properties[idx].X = x
				}

			case xOapiExtraTags:
				if err := val.Decode(&s.ExtraTags); err != nil {
					return err
				}
			}
			return nil
		}(); err != nil {
			return nil, errors.Wrapf(err, "parse %q", key)
		}
	}

	return s, nil
}

// validateGoIdent checks that given ident is valid and is exported.
func validateGoIdent(ident string) error {
	// TODO(tdakkota): move to generator package?
	// 	For now, keep as part of parser to use user-friendly location errors

	switch {
	case !token.IsIdentifier(ident):
		return errors.Errorf("invalid Go identifier %q", ident)
	case !token.IsExported(ident):
		return errors.Errorf("identifier must be public, got %q", ident)
	default:
		return nil
	}
}

func (p *Parser) parseSchema(schema *RawSchema, ctx *jsonpointer.ResolveCtx, hook func(*Schema) *Schema) (_ *Schema, err error) {
	if schema == nil {
		return nil, nil
	}
	wrapField := func(field string, err error) error {
		if err != nil {
			err = errors.Wrap(err, field)
		}
		return p.wrapField(field, p.file(ctx), schema.Common.Locator, err)
	}

	validateMinMax := func(prop string, min, max *uint64) (rerr error) {
		if min == nil || max == nil {
			return nil
		}
		if *min > *max {
			msg := fmt.Sprintf("min%s (%d) is greater than max%s (%d)", prop, *min, prop, *max)
			ptr := schema.Common.Locator.Pointer(p.file(ctx))

			me := new(location.MultiError)
			me.ReportPtr(ptr.Field("min"+prop), msg)
			me.ReportPtr(ptr.Field("max"+prop), "")
			return me
		}
		return nil
	}

	if ref := schema.Ref; ref != "" {
		s, err := p.resolve(ref, ctx)
		if err != nil {
			return nil, wrapField("$ref", err)
		}
		return s, nil
	}

	if schema.Type == "" && p.inferTypes {
		switch {
		case len(schema.Default) > 0:
			schema.Type, err = inferJSONType(json.RawMessage(schema.Default))
			if err != nil {
				return nil, wrapField("default", err)
			}

		case len(schema.Enum) > 0:
			schema.Type, err = inferJSONType(schema.Enum[0])
			if err != nil {
				return nil, errors.Wrap(err, "infer enum type")
			}
		default:
			// Try to infer schema type from properties.
			switch {
			case len(schema.Properties) > 0 ||
				schema.AdditionalProperties != nil ||
				schema.PatternProperties != nil ||
				schema.MaxProperties != nil ||
				schema.MinProperties != nil:
				schema.Type = "object"

			case schema.Items != nil ||
				schema.UniqueItems ||
				schema.MaxItems != nil ||
				schema.MinItems != nil:
				schema.Type = "array"

			case schema.Maximum != nil ||
				schema.Minimum != nil ||
				schema.ExclusiveMinimum ||
				schema.ExclusiveMaximum || // FIXME(tdakkota): check for existence instead of true?
				schema.MultipleOf != nil:
				schema.Type = "number"

			case schema.MaxLength != nil ||
				schema.MinLength != nil ||
				schema.Pattern != "":
				schema.Type = "string"
			}
		}
	}

	typ, ok := map[string]SchemaType{
		"object":  Object,
		"array":   Array,
		"string":  String,
		"integer": Integer,
		"number":  Number,
		"boolean": Boolean,
		"null":    Null,
		"":        Empty,
	}[schema.Type]
	if !ok {
		err := errors.Errorf("unexpected schema type: %q", schema.Type)
		return nil, wrapField("type", err)
	}

	if schema.Type != "" {
		allowed := map[string]map[string]struct{}{
			"object": {
				"required":          {},
				"properties":        {},
				"patternProperties": {},
				"minProperties":     {},
				"maxProperties":     {},
			},
			"array": {
				"items":       {},
				"maxItems":    {},
				"minItems":    {},
				"uniqueItems": {},
			},
			"string": {
				"maxLength": {},
				"minLength": {},
				"pattern":   {},
			},
			"integer": {
				"multipleOf":       {},
				"maximum":          {},
				"minimum":          {},
				"exclusiveMaximum": {},
				"exclusiveMinimum": {},
			},
			"number": {
				"multipleOf":       {},
				"maximum":          {},
				"minimum":          {},
				"exclusiveMaximum": {},
				"exclusiveMinimum": {},
			},
			"boolean": {},
			"null":    {},
		}

		for _, fset := range allowed {
			// Generic fields.
			for _, f := range []string{
				"type", "enum", "nullable", "format", "default",
				"oneOf", "anyOf", "allOf", "discriminator",
				"description", "example", "examples", "deprecated",
				"additionalProperties", "xml",
			} {
				fset[f] = struct{}{}
			}
		}

		if allowedFields, ok := allowed[schema.Type]; ok {
			fields, err := getRawSchemaFields(schema)
			if err != nil {
				return nil, err
			}

			for _, field := range fields {
				// Allow extra fields.
				if strings.HasPrefix(strings.ToLower(field), "x-") {
					continue
				}

				if _, ok := allowedFields[field]; !ok {
					return nil, wrapField(field, errors.Errorf("unexpected field for type %q", schema.Type))
				}
			}
		}
	}

	s := hook(&Schema{
		Type:     typ,
		Format:   schema.Format,
		Nullable: typ == Null,
		Required: slices.Clone(schema.Required),
		// Object validators
		MaxProperties: schema.MaxProperties,
		MinProperties: schema.MinProperties,
		// Array validators
		MinItems:    schema.MinItems,
		MaxItems:    schema.MaxItems,
		UniqueItems: schema.UniqueItems,
		// Number validators
		Minimum:          schema.Minimum,
		Maximum:          schema.Maximum,
		ExclusiveMinimum: schema.ExclusiveMinimum,
		ExclusiveMaximum: schema.ExclusiveMaximum,
		MultipleOf:       schema.MultipleOf,
		// String validators
		MaxLength: schema.MaxLength,
		MinLength: schema.MinLength,
		Pattern:   schema.Pattern,
	})

	switch {
	case len(schema.OneOf) > 0:
		s.OneOf, err = p.parseMany(schema.OneOf, schema.Common.Locator, ctx)
		if err != nil {
			return nil, wrapField("oneOf", err)
		}
	case len(schema.AnyOf) > 0:
		s.AnyOf, err = p.parseMany(schema.AnyOf, schema.Common.Locator, ctx)
		if err != nil {
			return nil, wrapField("anyOf", err)
		}
	case len(schema.AllOf) > 0:
		s.AllOf, err = p.parseMany(schema.AllOf, schema.Common.Locator, ctx)
		if err != nil {
			return nil, wrapField("allOf", err)
		}
	}

	// Object properties
	{
		if err := validateMinMax(
			"Properties",
			schema.MinProperties,
			schema.MaxProperties,
		); err != nil {
			return nil, err
		}

		if ap := schema.AdditionalProperties; ap != nil {
			var additional bool
			if val := ap.Bool; val != nil {
				additional = *val
			} else {
				additional = true
				if schema.Items != nil {
					ptr := schema.Common.Locator.Pointer(p.file(ctx))
					me := new(location.MultiError)
					me.ReportPtr(ptr.Field("additionalProperties"), "both additionalProperties and items fields are set")
					me.ReportPtr(ptr.Field("items"), "")
					return nil, me
				}

				s.Item, err = p.parse(&ap.Schema, ctx)
				if err != nil {
					return nil, wrapField("additionalProperties", err)
				}
			}
			s.AdditionalProperties = &additional
		}

		if pp := schema.PatternProperties; len(pp) > 0 {
			ppLoc := schema.Common.Locator.Field("patternProperties")

			patterns := make([]PatternProperty, len(pp))
			for idx, prop := range pp {
				pattern := prop.Pattern
				r, err := ogenregex.Compile(pattern)
				if err != nil {
					loc := ppLoc.Key(pattern)
					err := errors.Wrapf(err, "compile pattern %q", pattern)
					return nil, p.wrapLocation(p.file(ctx), loc, err)
				}
				sch, err := p.parse(prop.Schema, ctx)
				if err != nil {
					err := errors.Wrapf(err, "pattern schema %q", pattern)
					return nil, p.wrapField(pattern, p.file(ctx), ppLoc, err)
				}
				patterns[idx] = PatternProperty{
					Pattern: r,
					Schema:  sch,
				}
			}
			s.PatternProperties = patterns
		}

		propsLoc := schema.Common.Locator.Field("properties")
		for _, propSpec := range schema.Properties {
			prop, err := p.parse(propSpec.Schema, ctx)
			if err != nil {
				err := errors.Wrapf(err, "property %q", propSpec.Name)
				return nil, p.wrapField(propSpec.Name, p.file(ctx), propsLoc, err)
			}

			var description string
			if s := propSpec.Schema; s != nil {
				description = s.Description
			}

			required := slices.Contains(schema.Required, propSpec.Name)
			s.Properties = append(s.Properties, Property{
				Name:        propSpec.Name,
				Description: description,
				Schema:      prop,
				Required:    required,
			})
		}
	}

	// Array properties
	{
		if err := validateMinMax(
			"Items",
			schema.MinItems,
			schema.MaxItems,
		); err != nil {
			return nil, err
		}

		if items := schema.Items; items != nil {
			if item := items.Item; item != nil {
				s.Item, err = p.parse(items.Item, ctx)
				if err != nil {
					return nil, wrapField("items", err)
				}
			} else {
				itemsLoc := s.Locator.Field("items")
				if len(items.Items) == 0 {
					err := errors.New("array is empty")
					return nil, wrapField("items", err)
				}
				s.Items, err = p.parseMany(items.Items, itemsLoc, ctx)
				if err != nil {
					return nil, wrapField("items", err)
				}
			}
		}
	}

	// Integer, Number properties
	{
		if mul := schema.MultipleOf; len(mul) > 0 {
			if err := func() error {
				rat := new(big.Rat)
				if err := rat.UnmarshalText(mul); err != nil {
					return errors.Wrapf(err, "invalid number %q", mul)
				}
				// The value of "multipleOf" MUST be a number, strictly greater than 0.
				if rat.Sign() != 1 {
					return errors.Errorf("invalid multipleOf value %q", mul)
				}
				return nil
			}(); err != nil {
				return nil, wrapField("multipleOf", err)
			}
		}
	}

	// String properties
	{
		if err := validateMinMax(
			"Length",
			schema.MinLength,
			schema.MaxLength,
		); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (p *Parser) parseMany(schemas []*RawSchema, loc location.Locator, ctx *jsonpointer.ResolveCtx) ([]*Schema, error) {
	result := make([]*Schema, 0, len(schemas))
	for i, schema := range schemas {
		s, err := p.parse(schema, ctx)
		if err != nil {
			return nil, p.wrapLocation(p.file(ctx), loc.Index(i), err)
		}
		result = append(result, s)
	}

	return result, nil
}

func (p *Parser) extendInfo(schema *RawSchema, s *Schema, file location.File) *Schema {
	s.ContentEncoding = schema.ContentEncoding
	s.ContentMediaType = schema.ContentMediaType
	s.Summary = schema.Summary
	s.Description = schema.Description
	s.Deprecated = schema.Deprecated
	s.AddExample(schema.Example)

	// Nullable enums will be handled later.
	if len(s.Enum) < 1 {
		s.Nullable = schema.Nullable
	}
	if x := schema.XML; x != nil {
		s.XML = &XML{
			Name:      x.Name,
			Namespace: x.Namespace,
			Prefix:    x.Prefix,
			Attribute: x.Attribute,
			Wrapped:   x.Wrapped,
		}
	}

	s.Pointer = schema.Common.Locator.Pointer(file)
	return s
}

func (p *Parser) parseDiscriminator(d *RawDiscriminator, ctx *jsonpointer.ResolveCtx) (_ *Discriminator, rerr error) {
	locator := d.Common.Locator
	defer func() {
		rerr = p.wrapLocation(ctx.File(), locator, rerr)
	}()

	mapping := make(map[string]*Schema, len(d.Mapping))
	for value, ref := range d.Mapping {
		locator := locator.Field("mapping").Field(value)

		// See https://github.com/OAI/OpenAPI-Specification/issues/2520#issuecomment-1139961158.
		var (
			s   *Schema
			err error
		)
		switch {
		case !strings.ContainsRune(ref, '#'):
			// JSON Reference usually contains a fragment, e.g. "#/components/schemas/Foo" or
			// "foo.json#/definitions/Bar", but this looks like a schema name.
			//
			// Try to find it in the components, if it is root spec.
			if s, err = p.resolve("#/components/schemas/"+ref, ctx); err == nil {
				break
			}
			// It seems there is no schema with such name, try to resolve as a plain JSON Reference.
			fallthrough
		default:
			s, err = p.resolve(ref, ctx)
		}
		if err != nil {
			err = errors.Wrap(err, "resolve mapping")
			return nil, p.wrapLocation(ctx.File(), locator, err)
		}
		mapping[value] = s
	}

	return &Discriminator{
		PropertyName: d.PropertyName,
		Mapping:      mapping,
		Pointer:      locator.Pointer(ctx.File()),
	}, nil
}
