package ogen

import (
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
		Deprecated:           s.Deprecated,
		ContentEncoding:      s.ContentEncoding,
		ContentMediaType:     s.ContentMediaType,
		Discriminator:        s.Discriminator.ToJSONSchema(),
		XML:                  s.XML.ToJSONSchema(),
		Example:              s.Example,
		Common:               s.Common,
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
func (p PatternProperties) ToJSONSchema() (result jsonschema.RawPatternProperties) {
	for _, val := range p {
		result = append(result, jsonschema.RawPatternProperty{
			Pattern: val.Pattern,
			Schema:  val.Schema.ToJSONSchema(),
		})
	}
	return result
}

// ToJSONSchema converts Items to jsonschema.RawItems.
func (p *Items) ToJSONSchema() *jsonschema.RawItems {
	if p == nil {
		return nil
	}

	if item := p.Item; item != nil {
		return &jsonschema.RawItems{
			Item: item.ToJSONSchema(),
		}
	}
	rawItems := make([]*jsonschema.RawSchema, len(p.Items))
	for i, item := range p.Items {
		rawItems[i] = item.ToJSONSchema()
	}
	return &jsonschema.RawItems{
		Items: rawItems,
	}
}

// ToJSONSchema converts Discriminator to jsonschema.RawDiscriminator.
func (d *Discriminator) ToJSONSchema() *jsonschema.RawDiscriminator {
	if d == nil {
		return nil
	}

	return &jsonschema.RawDiscriminator{
		PropertyName: d.PropertyName,
		Mapping:      d.Mapping,
		Common:       d.Common,
	}
}

// ToJSONSchema converts XML to jsonschema.XML.
func (d *XML) ToJSONSchema() *jsonschema.XML {
	if d == nil {
		return nil
	}

	return &jsonschema.XML{
		Name:      d.Name,
		Namespace: d.Namespace,
		Prefix:    d.Prefix,
		Attribute: d.Attribute,
		Wrapped:   d.Wrapped,
	}
}
