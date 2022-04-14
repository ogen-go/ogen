package jsonschema

import (
	"math/big"
	"regexp"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// ReferenceResolver resolves JSON schema references.
type ReferenceResolver interface {
	ResolveReference(ref string) (*RawSchema, error)
}

// Parser parses JSON schemas.
type Parser struct {
	resolver   ReferenceResolver
	refcache   map[string]*Schema
	inferTypes bool
}

func NewParser(s Settings) *Parser {
	s.setDefaults()
	return &Parser{
		resolver:   s.Resolver,
		refcache:   s.ReferenceCache,
		inferTypes: s.InferTypes,
	}
}

func (p *Parser) Parse(schema *RawSchema) (*Schema, error) {
	return p.parse(schema, resolveCtx{})
}

func (p *Parser) parse(schema *RawSchema, ctx resolveCtx) (*Schema, error) {
	return p.parse1(schema, ctx, func(s *Schema) *Schema {
		return p.extendInfo(schema, s)
	})
}

func (p *Parser) parse1(schema *RawSchema, ctx resolveCtx, hook func(*Schema) *Schema) (*Schema, error) {
	s, err := p.parseSchema(schema, ctx, hook)
	if err != nil {
		return nil, errors.Wrap(err, "parse schema")
	}

	if schema != nil && s != nil {
		if len(schema.Enum) > 0 {
			values, err := parseEnumValues(s, schema.Enum)
			if err != nil {
				return nil, errors.Wrap(err, "parse enum values")
			}
			s.Enum = values
			handleNullableEnum(s)
		}
		if d := schema.Default; len(d) > 0 {
			v, err := parseJSONValue(s, d)
			if err != nil {
				return nil, errors.Wrap(err, "parse default")
			}
			if v == nil && !s.Nullable {
				return nil, errors.New("unexpected default \"null\" value")
			}
			s.Default = v
			s.DefaultSet = true
		}
		if a, ok := schema.XAnnotations["x-ogen-name"]; ok {
			name, err := jx.DecodeBytes(a).Str()
			if err != nil {
				return nil, errors.Wrapf(err, "decode %q", a)
			}
			s.XOgenName = name
		}
	}

	return s, nil
}

func (p *Parser) parseSchema(schema *RawSchema, ctx resolveCtx, hook func(*Schema) *Schema) (*Schema, error) {
	if schema == nil {
		return nil, nil
	}

	if ref := schema.Ref; ref != "" {
		s, err := p.resolve(ref, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q", ref)
		}
		return s, nil
	}

	if d := schema.Default; p.inferTypes && schema.Type == "" && len(d) > 0 {
		typ, err := inferJSONType(d)
		if err != nil {
			return nil, errors.Wrap(err, "infer default type")
		}
		schema.Type = typ
	}

	switch {
	case len(schema.Enum) > 0:
		typ := schema.Type
		if p.inferTypes && typ == "" {
			inferred, err := inferJSONType(schema.Enum[0])
			if err != nil {
				return nil, errors.Wrap(err, "infer enum type")
			}
			typ = inferred
		}

		t, err := parseType(typ)
		if err != nil {
			return nil, errors.Wrap(err, "type")
		}

		return hook(&Schema{
			Type:   t,
			Format: schema.Format,
		}), nil
	case len(schema.OneOf) > 0:
		s := hook(&Schema{})

		schemas, err := p.parseMany(schema.OneOf, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "oneOf")
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

		schemas, err := p.parseMany(schema.AnyOf, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "anyOf")
		}
		s.AnyOf = schemas

		return s, nil
	case len(schema.AllOf) > 0:
		s := hook(&Schema{})

		schemas, err := p.parseMany(schema.AllOf, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "allOf")
		}
		s.AllOf = schemas

		return s, nil
	}

	typ := schema.Type
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
			return nil, errors.New("object cannot contain 'items' field")
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
			// TODO(tdakkota): handle additionalProperties: false
			s.AdditionalProperties = true
			if !ap.Bool {
				item, err := p.parse(&ap.Schema, ctx)
				if err != nil {
					return nil, errors.Wrapf(err, "additionalProperties")
				}
				s.Item = item
			}
		}

		if pp := schema.PatternProperties; len(pp) > 0 {
			patterns := make([]PatternProperty, len(pp))
			for idx, prop := range pp {
				pattern := prop.Pattern
				r, err := regexp.Compile(pattern)
				if err != nil {
					return nil, errors.Wrapf(err, "compile pattern %q", pattern)
				}
				sch, err := p.parse(prop.Schema, ctx)
				if err != nil {
					return nil, errors.Wrapf(err, "pattern schema %q", pattern)
				}
				patterns[idx] = PatternProperty{
					Pattern: r,
					Schema:  sch,
				}
			}
			s.PatternProperties = patterns
		}

		for _, propSpec := range schema.Properties {
			prop, err := p.parse(propSpec.Schema, ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "property %q", propSpec.Name)
			}

			s.Properties = append(s.Properties, Property{
				Name:        propSpec.Name,
				Description: propSpec.Schema.Description,
				Schema:      prop,
				Required:    required(propSpec.Name),
			})
		}
		return s, nil

	case "array":
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
			return nil, errors.New("array cannot contain properties")
		}

		item, err := p.parse(schema.Items, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "item")
		}

		array.Item = item
		return array, nil

	case "number", "integer":
		if mul := schema.MultipleOf; mul != nil {
			rat := new(big.Rat)
			if err := rat.UnmarshalText(mul); err != nil {
				return nil, errors.Wrapf(err, "invalid number %q", mul)
			}
			// The value of "multipleOf" MUST be a number, strictly greater than 0.
			if rat.Sign() != 1 {
				return nil, errors.Errorf("invalid multipleOf value %q", mul)
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
		return hook(&Schema{}), nil

	default:
		return nil, errors.Errorf("unexpected schema type: %q", schema.Type)
	}
}

func (p *Parser) parseMany(schemas []*RawSchema, ctx resolveCtx) ([]*Schema, error) {
	result := make([]*Schema, 0, len(schemas))
	for i, schema := range schemas {
		s, err := p.parse(schema, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}

		result = append(result, s)
	}

	return result, nil
}

type resolveCtx map[string]struct{}

func (p *Parser) resolve(ref string, ctx resolveCtx) (*Schema, error) {
	if s, ok := p.refcache[ref]; ok {
		return s, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.New("infinite recursion")
	}
	ctx[ref] = struct{}{}
	defer func() {
		// Drop the resolved ref to prevent false-positive infinite recursion detection.
		delete(ctx, ref)
	}()

	raw, err := p.resolver.ResolveReference(ref)
	if err != nil {
		return nil, errors.Wrap(err, "find schema")
	}

	return p.parse1(raw, ctx, func(s *Schema) *Schema {
		s.Ref = ref
		p.refcache[ref] = s
		return p.extendInfo(raw, s)
	})
}

func (p *Parser) extendInfo(schema *RawSchema, s *Schema) *Schema {
	s.Description = schema.Description
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
	return s
}

func parseType(typ string) (SchemaType, error) {
	mapping := map[string]SchemaType{
		"object":  Object,
		"array":   Array,
		"boolean": Boolean,
		"integer": Integer,
		"number":  Number,
		"string":  String,
	}

	t, ok := mapping[typ]
	if !ok {
		return SchemaType(""), errors.Errorf("unexpected type: %q", typ)
	}

	return t, nil
}
