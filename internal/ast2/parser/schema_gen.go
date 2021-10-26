package parser

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	ast "github.com/ogen-go/ogen/internal/ast2"
)

// schemaGen is used to convert openapi schemas into ast representation.
type schemaGen struct {
	// Spec is used to lookup for schema components
	// which is not in the 'refs' cache.
	spec *ogen.Spec

	// Global schema reference cache from *Generator (readonly).
	globalRefs map[string]*ast.Schema

	// Local schema reference cache.
	localRefs map[string]*ast.Schema
}

// Generate converts ogen.Schema into *ast.Schema.
//
// If ogen.Schema contains references to schema components,
// these referenced schemas will be saved in g.localRefs.
func (g *schemaGen) Generate(schema ogen.Schema) (*ast.Schema, error) {
	s, err := g.generate(schema, "")
	if err != nil {
		return nil, xerrors.Errorf("gen: %w", err)
	}

	return s, nil
}

func (g *schemaGen) generate(schema ogen.Schema, ref string) (*ast.Schema, error) {
	if ref := schema.Ref; ref != "" {
		s, err := g.ref(ref)
		if err != nil {
			return nil, xerrors.Errorf("'%s': %w", ref, err)
		}
		return s, nil
	}

	if err := g.validateTypeFormat(schema.Type, schema.Format); err != nil {
		return nil, err
	}

	onret := func(s *ast.Schema) *ast.Schema {
		if ref != "" {
			g.localRefs[ref] = s
		}
		return s
	}

	switch {
	case len(schema.Enum) > 0:
		values, err := g.parseEnumValues(ast.SchemaType(schema.Type), schema.Enum)
		if err != nil {
			return nil, err
		}

		return onret(&ast.Schema{
			Type:        ast.SchemaType(schema.Type),
			Format:      schema.Format,
			Description: schema.Description,
			Enum:        values,
		}), nil
	case len(schema.OneOf) > 0:
		var schemas []*ast.Schema
		for i, s := range schema.OneOf {
			schema, err := g.generate(s, "")
			if err != nil {
				return nil, xerrors.Errorf("oneOf: %d: %w", i, err)
			}

			schemas = append(schemas, schema)
		}

		return onret(&ast.Schema{
			OneOf:       schemas,
			Ref:         schema.Ref,
			Description: schema.Description,
		}), nil
	case len(schema.AnyOf) > 0:
		var schemas []*ast.Schema
		for i, s := range schema.AnyOf {
			schema, err := g.generate(s, "")
			if err != nil {
				return nil, xerrors.Errorf("anyOf: %d: %w", i, err)
			}

			schemas = append(schemas, schema)
		}

		return onret(&ast.Schema{
			AnyOf:       schemas,
			Ref:         schema.Ref,
			Description: schema.Description,
		}), nil
	case len(schema.AllOf) > 0:
		var schemas []*ast.Schema
		for i, s := range schema.AllOf {
			schema, err := g.generate(s, "")
			if err != nil {
				return nil, xerrors.Errorf("allOf: %d: %w", i, err)
			}

			schemas = append(schemas, schema)
		}

		return onret(&ast.Schema{
			AllOf:       schemas,
			Ref:         schema.Ref,
			Description: schema.Description,
		}), nil
	}

	switch schema.Type {
	case "object":
		if schema.Items != nil {
			return nil, xerrors.New("object cannot contain 'items' field")
		}
		optional := func(name string) bool {
			for _, p := range schema.Required {
				if p == name {
					return false
				}
			}
			return true
		}
		s := &ast.Schema{
			Type:          ast.Object,
			Description:   schema.Description,
			Ref:           ref,
			MinProperties: schema.MinProperties,
			MaxProperties: schema.MaxProperties,
		}
		if ref != "" {
			g.localRefs[ref] = s
		}

		for propName, propSchema := range schema.Properties {
			prop, err := g.generate(propSchema, "")
			if err != nil {
				return nil, xerrors.Errorf("%s: %w", propName, err)
			}

			s.Properties = append(s.Properties, ast.Property{
				Name:     propName,
				Schema:   prop,
				Optional: optional(propName),
			})
		}
		sort.SliceStable(s.Properties, func(i, j int) bool {
			return strings.Compare(s.Properties[i].Name, s.Properties[j].Name) < 0
		})
		return s, nil

	case "array":
		array := &ast.Schema{
			Type:        ast.Array,
			Description: schema.Description,
			Ref:         ref,
			MinItems:    schema.MinItems,
			MaxItems:    schema.MaxItems,
			UniqueItems: schema.UniqueItems,
		}
		if schema.Items == nil {
			// Fallback to string.
			array.Item = &ast.Schema{Type: ast.String}
			return array, nil
		}
		if len(schema.Properties) > 0 {
			return nil, xerrors.New("array cannot contain properties")
		}

		if ref != "" {
			g.localRefs[ref] = array
		}

		item, err := g.generate(*schema.Items, "")
		if err != nil {
			return nil, err
		}

		array.Item = item
		return array, nil

	case "number", "integer":
		return onret(&ast.Schema{
			Type:             ast.SchemaType(schema.Type),
			Format:           schema.Format,
			Description:      schema.Description,
			Ref:              ref,
			Minimum:          schema.Minimum,
			Maximum:          schema.Maximum,
			ExclusiveMinimum: schema.ExclusiveMinimum,
			ExclusiveMaximum: schema.ExclusiveMaximum,
			MultipleOf:       schema.MultipleOf,
		}), nil

	case "boolean":
		return onret(&ast.Schema{
			Type:        ast.Boolean,
			Format:      schema.Format,
			Description: schema.Description,
			Ref:         ref,
		}), nil

	case "string":
		return onret(&ast.Schema{
			Type:        ast.String,
			Format:      schema.Format,
			Description: schema.Description,
			Ref:         ref,
			MaxLength:   schema.MaxLength,
			Pattern:     schema.Pattern,
		}), nil

	case "":
		return onret(&ast.Schema{Type: ast.String}), nil

	default:
		return nil, xerrors.Errorf("unexpected schema type: '%s'", schema.Type)
	}
}

func (g *schemaGen) ref(ref string) (*ast.Schema, error) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, fmt.Errorf("invalid schema reference '%s'", ref)
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
		return nil, xerrors.Errorf("component by reference '%s' not found", ref)
	}

	return g.generate(component, ref)
}

func (g *schemaGen) validateTypeFormat(typ, format string) error {
	switch typ {
	case "object":
		switch format {
		case "":
			return nil
		default:
			return xerrors.Errorf("unexpected object format: '%s'", format)
		}
	case "array":
		switch format {
		case "":
			return nil
		default:
			return xerrors.Errorf("unexpected array format: '%s'", format)
		}
	case "integer":
		switch format {
		case "int32", "int64", "":
			return nil
		default:
			return xerrors.Errorf("unexpected integer format: '%s'", format)
		}
	case "number":
		switch format {
		case "float", "double", "":
			return nil
		default:
			return xerrors.Errorf("unexpected number format: '%s'", format)
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
			// return nil, xerrors.Errorf("unexpected string format: '%s'", format)
			return nil
		}
	case "boolean":
		switch format {
		case "":
			return nil
		default:
			return xerrors.Errorf("unexpected bool format: '%s'", format)
		}
	default:
		return xerrors.Errorf("unexpected type: '%s'", typ)
	}
}
