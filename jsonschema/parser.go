// Package jsonschema contains parser for JSON Schema.
package jsonschema

import (
	"encoding/json"
	"math/big"
	"regexp"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/internal/location"
	ogenjson "github.com/ogen-go/ogen/json"
)

// Parser parses JSON schemas.
type Parser struct {
	external ExternalResolver
	schemas  map[string]ReferenceResolver
	refcache map[jsonpointer.RefKey]*Schema

	filename string // optional, used for error messages

	inferTypes bool
}

// NewParser creates new Parser.
func NewParser(s Settings) *Parser {
	s.setDefaults()
	return &Parser{
		external: s.External,
		schemas: map[string]ReferenceResolver{
			"": s.Resolver,
		},
		refcache:   map[jsonpointer.RefKey]*Schema{},
		filename:   s.Filename,
		inferTypes: s.InferTypes,
	}
}

// Parse parses given RawSchema and returns parsed Schema.
func (p *Parser) Parse(schema *RawSchema) (*Schema, error) {
	return p.ParseWithContext(schema, jsonpointer.DefaultCtx())
}

// ParseWithContext parses given RawSchema and returns parsed Schema.
func (p *Parser) ParseWithContext(schema *RawSchema, ctx *jsonpointer.ResolveCtx) (*Schema, error) {
	return p.parse(schema, ctx)
}

// Resolve resolves Schema by given ref.
func (p *Parser) Resolve(ref string, ctx *jsonpointer.ResolveCtx) (*Schema, error) {
	return p.resolve(ref, ctx)
}

func (p *Parser) parse(schema *RawSchema, ctx *jsonpointer.ResolveCtx) (_ *Schema, rerr error) {
	if schema != nil {
		defer func() {
			rerr = p.wrapLocation(ctx.LastLoc(), schema.Common.Locator, rerr)
		}()
	}
	return p.parse1(schema, ctx, func(s *Schema) *Schema {
		return p.extendInfo(schema, s)
	})
}

func (p *Parser) parse1(schema *RawSchema, ctx *jsonpointer.ResolveCtx, hook func(*Schema) *Schema) (*Schema, error) {
	s, err := p.parseSchema(schema, ctx, hook)
	if err != nil {
		return nil, errors.Wrap(err, "parse schema")
	}

	if schema != nil && s != nil {
		if enum := schema.Enum; len(enum) > 0 {
			loc := schema.Common.Locator.Field("enum")
			for i := range enum {
				for j := range enum {
					if i == j {
						continue
					}
					a, b := enum[i], enum[j]
					if ok, _ := ogenjson.Equal(a, b); ok {
						loc := loc.Index(i)
						err := errors.Errorf("duplicate enum value: %q", a)
						return nil, p.wrapLocation(ctx.LastLoc(), loc, err)
					}
				}
			}

			values, err := parseEnumValues(s, enum)
			if err != nil {
				err := errors.Wrap(err, "parse enum values")
				return nil, p.wrapLocation(ctx.LastLoc(), loc, err)
			}
			s.Enum = values
			handleNullableEnum(s)
		}
		if d := schema.Default; len(d) > 0 {
			if err := func() error {
				v, err := parseJSONValue(s, json.RawMessage(d))
				if err != nil {
					return err
				}

				if v == nil && !s.Nullable {
					return errors.New("unexpected default \"null\" value")
				}

				s.Default = v
				s.DefaultSet = true
				return nil
			}(); err != nil {
				err := errors.Wrap(err, "parse default")
				return nil, p.wrapField("default", ctx.LastLoc(), schema.Common.Locator, err)
			}
		}
		if a, ok := schema.Common.Extensions["x-ogen-name"]; ok {
			if err := a.Decode(&s.XOgenName); err != nil {
				return nil, errors.Wrap(err, "parse x-ogen-name")
			}
		}
	}

	return s, nil
}

func (p *Parser) parseSchema(schema *RawSchema, ctx *jsonpointer.ResolveCtx, hook func(*Schema) *Schema) (*Schema, error) {
	if schema == nil {
		return nil, nil
	}
	wrapField := func(field string, err error) error {
		return p.wrapField(field, ctx.LastLoc(), schema.Common.Locator, err)
	}

	validateMinMax := func(prop string, min, max *uint64) (rerr error) {
		if min == nil || max == nil {
			return nil
		}
		defer func() {
			if rerr != nil {
				rerr = wrapField("min"+prop, rerr)
			}
		}()

		if *min > *max {
			return errors.Errorf("min%s (%d) is greater than max%s (%d)", prop, *min, prop, *max)
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

	if d := schema.Default; p.inferTypes && schema.Type == "" && len(d) > 0 {
		typ, err := inferJSONType(json.RawMessage(d))
		if err != nil {
			return nil, wrapField("default", err)
		}
		schema.Type = typ
	}

	typ := schema.Type
	switch {
	case len(schema.Enum) > 0:
		if p.inferTypes && typ == "" {
			inferred, err := inferJSONType(schema.Enum[0])
			if err != nil {
				return nil, errors.Wrap(err, "infer enum type")
			}
			typ = inferred
		}
	case len(schema.OneOf) > 0:
		s := hook(&Schema{})

		schemas, err := p.parseMany(schema.OneOf, schema.Common.Locator, ctx)
		if err != nil {
			return nil, wrapField("oneOf", err)
		}
		s.OneOf = schemas

		return s, nil
	case len(schema.AnyOf) > 0:
		s := hook(&Schema{
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

		schemas, err := p.parseMany(schema.AnyOf, schema.Common.Locator, ctx)
		if err != nil {
			return nil, wrapField("anyOf", err)
		}
		s.AnyOf = schemas

		return s, nil
	case len(schema.AllOf) > 0:
		s := hook(&Schema{})

		schemas, err := p.parseMany(schema.AllOf, schema.Common.Locator, ctx)
		if err != nil {
			return nil, wrapField("allOf", err)
		}
		s.AllOf = schemas

		return s, nil
	}

	// Try to infer schema type from properties.
	if p.inferTypes && typ == "" {
		switch {
		case len(schema.Properties) > 0 ||
			schema.AdditionalProperties != nil ||
			schema.PatternProperties != nil ||
			schema.MaxProperties != nil ||
			schema.MinProperties != nil:
			typ = "object"

		case schema.Items != nil ||
			schema.UniqueItems ||
			schema.MaxItems != nil ||
			schema.MinItems != nil:
			typ = "array"

		case schema.Maximum != nil ||
			schema.Minimum != nil ||
			schema.ExclusiveMinimum ||
			schema.ExclusiveMaximum || // FIXME(tdakkota): check for existence instead of true?
			schema.MultipleOf != nil:
			typ = "number"

		case schema.MaxLength != nil ||
			schema.MinLength != nil ||
			schema.Pattern != "":
			typ = "string"
		}
	}

	switch typ {
	case "object":
		if schema.Items != nil {
			err := errors.New("object cannot contain 'items' field")
			return nil, wrapField("items", err)
		}
		if err := validateMinMax(
			"Properties",
			schema.MinProperties,
			schema.MaxProperties,
		); err != nil {
			return nil, err
		}

		required := func(name string) bool {
			for _, p := range schema.Required {
				if p == name {
					return true
				}
			}
			return false
		}

		s := hook(&Schema{
			Type:          Object,
			MaxProperties: schema.MaxProperties,
			MinProperties: schema.MinProperties,
		})

		if ap := schema.AdditionalProperties; ap != nil {
			var additional bool
			if val := ap.Bool; val != nil {
				additional = *val
			} else {
				additional = true
				item, err := p.parse(&ap.Schema, ctx)
				if err != nil {
					return nil, wrapField("additionalProperties", err)
				}
				s.Item = item
			}
			s.AdditionalProperties = &additional
		}

		if pp := schema.PatternProperties; len(pp) > 0 {
			ppLoc := schema.Common.Locator.Field("patternProperties")

			patterns := make([]PatternProperty, len(pp))
			for idx, prop := range pp {
				pattern := prop.Pattern
				r, err := regexp.Compile(pattern)
				if err != nil {
					loc := ppLoc.Key(pattern)
					err := errors.Wrapf(err, "compile pattern %q", pattern)
					return nil, p.wrapLocation(ctx.LastLoc(), loc, err)
				}
				sch, err := p.parse(prop.Schema, ctx)
				if err != nil {
					err := errors.Wrapf(err, "pattern schema %q", pattern)
					return nil, p.wrapField(pattern, ctx.LastLoc(), ppLoc, err)
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
				return nil, p.wrapField(propSpec.Name, ctx.LastLoc(), propsLoc, err)
			}

			var description string
			if s := propSpec.Schema; s != nil {
				description = s.Description
			}

			s.Properties = append(s.Properties, Property{
				Name:        propSpec.Name,
				Description: description,
				Schema:      prop,
				Required:    required(propSpec.Name),
			})
		}
		return s, nil

	case "array":
		if err := validateMinMax(
			"Items",
			schema.MinItems,
			schema.MaxItems,
		); err != nil {
			return nil, err
		}

		array := hook(&Schema{
			Type:        Array,
			MinItems:    schema.MinItems,
			MaxItems:    schema.MaxItems,
			UniqueItems: schema.UniqueItems,
		})

		if schema.Items == nil {
			// Keep items nil, we will use ir.Any for it.
			return array, nil
		}
		if len(schema.Properties) > 0 {
			err := errors.New("array cannot contain properties")
			return nil, wrapField("properties", err)
		}

		item, err := p.parse(schema.Items, ctx)
		if err != nil {
			return nil, wrapField("items", err)
		}

		array.Item = item
		return array, nil

	case "number", "integer":
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

		return hook(&Schema{
			Type:             SchemaType(schema.Type),
			Format:           schema.Format,
			Minimum:          schema.Minimum,
			Maximum:          schema.Maximum,
			ExclusiveMinimum: schema.ExclusiveMinimum,
			ExclusiveMaximum: schema.ExclusiveMaximum,
			MultipleOf:       schema.MultipleOf,
		}), nil

	case "boolean":
		return hook(&Schema{
			Type:   Boolean,
			Format: schema.Format,
		}), nil

	case "string":
		if err := validateMinMax(
			"Length",
			schema.MinLength,
			schema.MaxLength,
		); err != nil {
			return nil, err
		}

		return hook(&Schema{
			Type:      String,
			Format:    schema.Format,
			MaxLength: schema.MaxLength,
			MinLength: schema.MinLength,
			Pattern:   schema.Pattern,
		}), nil

	case "null":
		return hook(&Schema{
			Type:     Null,
			Nullable: true,
		}), nil

	case "":
		return hook(&Schema{
			Format: schema.Format,
		}), nil

	default:
		err := errors.Errorf("unexpected schema type: %q", schema.Type)
		return nil, wrapField("type", err)
	}
}

func (p *Parser) parseMany(schemas []*RawSchema, loc location.Locator, ctx *jsonpointer.ResolveCtx) ([]*Schema, error) {
	result := make([]*Schema, 0, len(schemas))
	for i, schema := range schemas {
		s, err := p.parse(schema, ctx)
		if err != nil {
			return nil, p.wrapLocation(ctx.LastLoc(), loc.Index(i), err)
		}
		result = append(result, s)
	}

	return result, nil
}

func (p *Parser) extendInfo(schema *RawSchema, s *Schema) *Schema {
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
	if d := schema.Discriminator; d != nil {
		s.Discriminator = &Discriminator{
			PropertyName: d.PropertyName,
			Mapping:      d.Mapping,
		}
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

	s.Locator = schema.Common.Locator
	return s
}
