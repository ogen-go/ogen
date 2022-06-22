package ogen

import (
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/jsonschema"
)

// ToJSONSchema converts Schema to jsonschema.Schema.
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
		Summary:              s.Summary,
		Description:          s.Description,
		Type:                 s.Type,
		Format:               s.Format,
		Properties:           s.Properties.ToJSONSchema(),
		AdditionalProperties: s.AdditionalProperties.ToJSONSchema(),
		PatternProperties:    s.PatternProperties.ToJSONSchema(),
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
		ContentEncoding:      s.ContentEncoding,
		ContentMediaType:     s.ContentMediaType,
		XAnnotations:         map[string]jx.Raw{},
	}
}

// ToJSONSchema converts Properties to jsonschema.RawProperties.
func (p Properties) ToJSONSchema() jsonschema.RawProperties {
	result := make([]jsonschema.RawProperty, 0, len(p))
	for _, prop := range p {
		result = append(result, prop.ToJSONSchema())
	}
	return result
}

// ToJSONSchema converts Property to jsonschema.Property.
func (p Property) ToJSONSchema() jsonschema.RawProperty {
	return jsonschema.RawProperty{
		Name:   p.Name,
		Schema: p.Schema.ToJSONSchema(),
	}
}

// ToJSONSchema converts AdditionalProperties to jsonschema.AdditionalProperties.
func (p *AdditionalProperties) ToJSONSchema() *jsonschema.AdditionalProperties {
	if p == nil {
		return nil
	}

	var val *bool
	if p.Bool != nil {
		val = new(bool)
		*val = *p.Bool
	}
	return &jsonschema.AdditionalProperties{
		Bool:   val,
		Schema: *p.Schema.ToJSONSchema(),
	}
}

// ToJSONSchema converts PatternProperties to jsonschema.RawPatternProperties.
func (r PatternProperties) ToJSONSchema() (result jsonschema.RawPatternProperties) {
	for _, val := range r {
		result = append(result, jsonschema.RawPatternProperty{
			Pattern: val.Pattern,
			Schema:  val.Schema.ToJSONSchema(),
		})
	}
	return result
}

// ToJSONSchema converts Discriminator to jsonschema.Discriminator.
func (d *Discriminator) ToJSONSchema() *jsonschema.Discriminator {
	if d == nil {
		return nil
	}

	return &jsonschema.Discriminator{
		PropertyName: d.PropertyName,
		Mapping:      d.Mapping,
	}
}
