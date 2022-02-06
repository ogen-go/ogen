package ogen

import "github.com/ogen-go/ogen/internal/jsonschema"

func (s *Schema) ToJSONSchema() *jsonschema.RawSchema {
	if s == nil {
		return nil
	}

	convertMany := func(schemas []*Schema) []*jsonschema.RawSchema {
		result := make([]*jsonschema.RawSchema, 0, len(schemas))
		for _, s := range schemas {
			result = append(result, s.ToJSONSchema())
		}
		return result
	}

	return &jsonschema.RawSchema{
		Ref:                  s.Ref,
		Description:          s.Description,
		Type:                 s.Type,
		Format:               s.Format,
		Properties:           s.Properties.ToJSONSchema(),
		AdditionalProperties: s.AdditionalProperties.ToJSONSchema(),
		Required:             s.Required,
		Items:                s.Items.ToJSONSchema(),
		Nullable:             s.Nullable,
		AllOf:                convertMany(s.AllOf),
		OneOf:                convertMany(s.OneOf),
		AnyOf:                convertMany(s.AnyOf),
		Discriminator:        s.Discriminator.ToJSONSchema(),
		Enum:                 s.Enum,
		MultipleOf:           s.MultipleOf,
		Maximum:              s.Maximum,
		ExclusiveMaximum:     s.ExclusiveMaximum,
		Minimum:              s.Minimum,
		ExclusiveMinimum:     s.ExclusiveMinimum,
		MaxLength:            s.MaxLength,
		MinLength:            s.MinLength,
		Pattern:              s.Pattern,
		MaxItems:             s.MaxItems,
		MinItems:             s.MinItems,
		UniqueItems:          s.UniqueItems,
		MaxProperties:        s.MaxProperties,
		MinProperties:        s.MinProperties,
		Default:              s.Default,
		Example:              s.Example,
		Deprecated:           s.Deprecated,
	}
}

func (p Properties) ToJSONSchema() jsonschema.RawProperties {
	result := make([]jsonschema.RawProperty, 0, len(p))
	for _, prop := range p {
		result = append(result, prop.ToJSONSchema())
	}
	return result
}

func (p Property) ToJSONSchema() jsonschema.RawProperty {
	return jsonschema.RawProperty{
		Name:   p.Name,
		Schema: p.Schema.ToJSONSchema(),
	}
}

func (p *AdditionalProperties) ToJSONSchema() *jsonschema.AdditionalProperties {
	if p == nil {
		return nil
	}

	return &jsonschema.AdditionalProperties{
		Bool:   p.Bool,
		Schema: *p.Schema.ToJSONSchema(),
	}
}

func (d *Discriminator) ToJSONSchema() *jsonschema.Discriminator {
	if d == nil {
		return nil
	}

	return &jsonschema.Discriminator{
		PropertyName: d.PropertyName,
		Mapping:      d.Mapping,
	}
}
