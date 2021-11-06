package parser

import (
	"sort"
	"strings"

	"github.com/ogen-go/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

// schemaGen is used to convert openapi schemas into oas representation.
type schemaGen struct {
	// Spec is used to lookup for schema components
	// which is not in the 'refs' cache.
	spec *ogen.Spec

	// Global schema reference cache from *Generator (readonly).
	globalRefs map[string]*oas.Schema

	// Local schema reference cache.
	localRefs map[string]*oas.Schema
}

// Generate converts ogen.Schema into *oas.Schema.
//
// If ogen.Schema contains references to schema components,
// these referenced schemas will be saved in g.localRefs.
func (g *schemaGen) Generate(schema ogen.Schema) (*oas.Schema, error) {
	s, err := g.generate(schema, "")
	if err != nil {
		return nil, errors.Wrap(err, "gen")
	}

	return s, nil
}

func (g *schemaGen) generate(schema ogen.Schema, ref string) (*oas.Schema, error) {
	if ref := schema.Ref; ref != "" {
		s, err := g.ref(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "ref %q", ref)
		}
		return s, nil
	}

	// extendInfo extends provided schema with common OpenAPI fields.
	// Must be called on each success return.
	extendInfo := func(s *oas.Schema) *oas.Schema {
		s.Ref = ref
		s.Description = schema.Description
		s.Nullable = schema.Nullable
		if d := schema.Discriminator; d != nil {
			s.Discriminator = &oas.Discriminator{
				PropertyName: d.PropertyName,
				Mapping:      d.Mapping,
			}
		}
		if ref != "" {
			g.localRefs[ref] = s
		}
		return s
	}

	switch {
	case len(schema.Enum) > 0:
		if err := validateTypeFormat(schema.Type, schema.Format); err != nil {
			return nil, errors.Wrap(err, "validate format")
		}

		values, err := parseEnumValues(oas.SchemaType(schema.Type), schema.Enum)
		if err != nil {
			return nil, errors.Wrap(err, "parse enum")
		}

		return extendInfo(&oas.Schema{
			Type:   oas.SchemaType(schema.Type),
			Format: oas.Format(schema.Format),
			Enum:   values,
		}), nil
	case len(schema.OneOf) > 0:
		var schemas []*oas.Schema
		for i, s := range schema.OneOf {
			schema, err := g.generate(s, "")
			if err != nil {
				return nil, errors.Wrapf(err, "oneOf: %d", i)
			}

			schemas = append(schemas, schema)
		}

		return extendInfo(&oas.Schema{OneOf: schemas}), nil
	case len(schema.AnyOf) > 0:
		var schemas []*oas.Schema
		for i, s := range schema.AnyOf {
			schema, err := g.generate(s, "")
			if err != nil {
				return nil, errors.Wrapf(err, "anyOf: %d", i)
			}

			schemas = append(schemas, schema)
		}

		return extendInfo(&oas.Schema{AnyOf: schemas}), nil
	case len(schema.AllOf) > 0:
		var schemas []*oas.Schema
		for i, s := range schema.AllOf {
			schema, err := g.generate(s, "")
			if err != nil {
				return nil, errors.Wrapf(err, "allOf: %d", i)
			}

			schemas = append(schemas, schema)
		}

		return extendInfo(&oas.Schema{AllOf: schemas}), nil
	}

	switch schema.Type {
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
		s := extendInfo(&oas.Schema{
			Type:          oas.Object,
			MinProperties: schema.MinProperties,
			MaxProperties: schema.MaxProperties,
		})

		// Ensure that order is stable.
		propKeys := schema.XPropertiesOrder
		for _, k := range schema.XPropertiesOrder {
			if _, ok := schema.Properties[k]; ok {
				continue
			}
			return nil, errors.Errorf("invalid x-properties-order: missing %s", k)
		}
		if len(propKeys) == 0 {
			for k := range schema.Properties {
				propKeys = append(propKeys, k)
			}
			sort.Strings(propKeys)
		}

		for _, propName := range propKeys {
			propSchema := schema.Properties[propName]
			prop, err := g.generate(propSchema, "")
			if err != nil {
				return nil, errors.Wrapf(err, "%s", propName)
			}

			s.Properties = append(s.Properties, oas.Property{
				Name:     propName,
				Schema:   prop,
				Required: required(propName),
			})
		}
		return s, nil

	case "array":
		array := extendInfo(&oas.Schema{
			Type:        oas.Array,
			MinItems:    schema.MinItems,
			MaxItems:    schema.MaxItems,
			UniqueItems: schema.UniqueItems,
		})
		if schema.Items == nil {
			// Fallback to string.
			array.Item = &oas.Schema{Type: oas.String}
			return array, nil
		}
		if len(schema.Properties) > 0 {
			return nil, errors.New("array cannot contain properties")
		}

		item, err := g.generate(*schema.Items, "")
		if err != nil {
			return nil, errors.Wrap(err, "item")
		}

		array.Item = item
		return array, nil

	case "number", "integer":
		if err := validateTypeFormat(schema.Type, schema.Format); err != nil {
			return nil, errors.Wrap(err, "validate format")
		}

		return extendInfo(&oas.Schema{
			Type:             oas.SchemaType(schema.Type),
			Format:           oas.Format(schema.Format),
			Minimum:          schema.Minimum,
			Maximum:          schema.Maximum,
			ExclusiveMinimum: schema.ExclusiveMinimum,
			ExclusiveMaximum: schema.ExclusiveMaximum,
			MultipleOf:       schema.MultipleOf,
		}), nil

	case "boolean":
		if err := validateTypeFormat(schema.Type, schema.Format); err != nil {
			return nil, errors.Wrap(err, "validate format")
		}

		return extendInfo(&oas.Schema{
			Type:   oas.Boolean,
			Format: oas.Format(schema.Format),
		}), nil

	case "string":
		if err := validateTypeFormat(schema.Type, schema.Format); err != nil {
			return nil, errors.Wrap(err, "validate format")
		}

		return extendInfo(&oas.Schema{
			Type:      oas.String,
			Format:    oas.Format(schema.Format),
			MaxLength: schema.MaxLength,
			MinLength: schema.MinLength,
			Pattern:   schema.Pattern,
		}), nil

	case "":
		return extendInfo(&oas.Schema{Type: oas.String}), nil

	default:
		return nil, errors.Errorf("unexpected schema type: %q", schema.Type)
	}
}

func (g *schemaGen) ref(ref string) (*oas.Schema, error) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid schema reference %q", ref)
	}

	if s, ok := g.globalRefs[ref]; ok {
		return s, nil
	}

	if s, ok := g.localRefs[ref]; ok {
		return s, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	component, found := g.spec.Components.Schemas[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	return g.generate(component, ref)
}

func validateTypeFormat(typ, format string) error {
	switch typ {
	case "object":
		switch format {
		case "":
			return nil
		default:
			return errors.Errorf("unexpected object format: %q", format)
		}
	case "array":
		switch format {
		case "":
			return nil
		default:
			return errors.Errorf("unexpected array format: %q", format)
		}
	case "integer":
		switch format {
		case "int32", "int64", "":
			return nil
		default:
			return errors.Errorf("unexpected integer format: %q", format)
		}
	case "number":
		switch format {
		case "float", "double", "":
			return nil
		default:
			return errors.Errorf("unexpected number format: %q", format)
		}
	case "string":
		switch format {
		case "byte":
			return nil
		case "date-time", "time", "date":
			return nil
		case "duration":
			return nil
		case "uuid":
			return nil
		case "ipv4", "ipv6", "ip":
			return nil
		case "uri":
			return nil
		case "password", "":
			return nil
		default:
			// Tools that do not recognize a specific format MAY default
			// back to the type alone, as if the format is not specified.
			return nil
		}
	case "boolean":
		switch format {
		case "":
			return nil
		default:
			return errors.Errorf("unexpected bool format: %q", format)
		}
	default:
		return errors.Errorf("unexpected type: %q", typ)
	}
}
