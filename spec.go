package ogen

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// Spec is the root document object of the OpenAPI document.
type Spec struct {
	// This string MUST be the semantic version number
	// of the OpenAPI Specification version that the OpenAPI document uses.
	OpenAPI    string      `json:"openapi"`
	Info       Info        `json:"info"`
	Servers    []Server    `json:"servers,omitempty"`
	Paths      Paths       `json:"paths,omitempty"`
	Components *Components `json:"components,omitempty"`

	// A list of tags used by the specification with additional metadata.
	// The order of the tags can be used to reflect on their order by the parsing
	// tools. Not all tags that are used by the Operation Object must be declared.
	// The tags that are not declared MAY be organized randomly or based on the tools' logic.
	// Each tag name in the list MUST be unique.
	Tags []Tag `json:"tags,omitempty"`
}

// Init components of schema.
func (s *Spec) Init() {
	if s.Components == nil {
		s.Components = &Components{}
	}

	c := s.Components
	if c.Schemas == nil {
		c.Schemas = make(map[string]*Schema)
	}
	if c.Responses == nil {
		c.Responses = make(map[string]*Response)
	}
	if c.Parameters == nil {
		c.Parameters = make(map[string]*Parameter)
	}
	if c.RequestBodies == nil {
		c.RequestBodies = make(map[string]*RequestBody)
	}
	if c.Examples == nil {
		c.Examples = make(map[string]*Example)
	}
}

// Example object.
//
// https://swagger.io/specification/#example-object
type Example struct {
	Ref           string          `json:"$ref,omitempty"` // ref object
	Summary       string          `json:"summary,omitempty"`
	Description   string          `json:"description,omitempty"`
	Value         json.RawMessage `json:"value,omitempty"`
	ExternalValue string          `json:"externalValue,omitempty"`
}

// Tag object.
//
// https://swagger.io/specification/#tag-object
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Info provides metadata about the API.
//
// The metadata MAY be used by the clients if needed,
// and MAY be presented in editing or documentation generation tools for convenience.
type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title"`
	// A short description of the API.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// A URL to the Terms of Service for the API. MUST be in the format of a URL.
	TermsOfService string `json:"termsOfService,omitempty"`
	// The contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty"`
	// The license information for the exposed API.
	License *License `json:"license,omitempty"`
	// REQUIRED. The version of the OpenAPI document.
	Version string `json:"version"`
}

// Contact information for the exposed API.
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License information for the exposed API.
type License struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Server represents a Server.
type Server struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// Components hold a set of reusable objects for different aspects of the OAS.
// All objects defined within the components object will have no effect on the API
// unless they are explicitly referenced from properties outside the components object.
type Components struct {
	Schemas       map[string]*Schema      `json:"schemas,omitempty"`
	Responses     map[string]*Response    `json:"responses,omitempty"`
	Parameters    map[string]*Parameter   `json:"parameters,omitempty"`
	Examples      map[string]*Example     `json:"examples,omitempty"`
	RequestBodies map[string]*RequestBody `json:"requestBodies,omitempty"`

	// Headers         map[string]Header          `json:"headers"`
	// SecuritySchemes map[string]SecuritySchema  `json:"securitySchemes"`
	// Links           map[string]Link            `json:"links"`
	// Callbacks       map[string]Callback        `json:"callback"`
}

// Paths holds the relative paths to the individual endpoints and their operations.
// The path is appended to the URL from the Server Object in order to construct the full URL.
// The Paths MAY be empty, due to ACL constraints.
type Paths map[string]*PathItem

// PathItem describes the operations available on a single path.
// A Path Item MAY be empty, due to ACL constraints.
// The path itself is still exposed to the documentation viewer,
// but they will not know which operations and parameters are available.
type PathItem struct {
	// Allows for an external definition of this path item.
	// The referenced structure MUST be in the format of a Path Item Object.
	// In case a Path Item Object field appears both
	// in the defined object and the referenced object, the behavior is undefined.
	Ref         string       `json:"$ref,omitempty"`
	Description string       `json:"description,omitempty"`
	Get         *Operation   `json:"get,omitempty"`
	Put         *Operation   `json:"put,omitempty"`
	Post        *Operation   `json:"post,omitempty"`
	Delete      *Operation   `json:"delete,omitempty"`
	Options     *Operation   `json:"options,omitempty"`
	Head        *Operation   `json:"head,omitempty"`
	Patch       *Operation   `json:"patch,omitempty"`
	Trace       *Operation   `json:"trace,omitempty"`
	Servers     []Server     `json:"servers,omitempty"`
	Parameters  []*Parameter `json:"parameters,omitempty"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	// A list of tags for API documentation control.
	// Tags can be used for logical grouping of operations by resources or any other qualifier.
	Tags        []string     `json:"tags,omitempty"`
	Summary     string       `json:"summary,omitempty"`
	Description string       `json:"description,omitempty"`
	OperationID string       `json:"operationId,omitempty"`
	Parameters  []*Parameter `json:"parameters,omitempty"`
	RequestBody *RequestBody `json:"requestBody,omitempty"`
	Responses   Responses    `json:"responses,omitempty"`
	Deprecated  bool         `json:"deprecated,omitempty"`
}

// Parameter describes a single operation parameter.
// A unique parameter is defined by a combination of a name and location.
type Parameter struct {
	Ref  string `json:"$ref,omitempty"`
	Name string `json:"name"`

	// The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`

	// Determines whether this parameter is mandatory.
	// If the parameter location is "path", this property is REQUIRED
	// and its value MUST be true.
	// Otherwise, the property MAY be included and its default value is false.
	Required bool `json:"required,omitempty"`

	// Specifies that a parameter is deprecated and SHOULD be transitioned out of usage.
	// Default value is false.
	Deprecated bool `json:"deprecated,omitempty"` // TODO: implement

	// For more complex scenarios, the content property can define the media type and schema of the parameter.
	// A parameter MUST contain either a schema property, or a content property, but not both.
	// When example or examples are provided in conjunction with the schema object,
	// the example MUST follow the prescribed serialization strategy for the parameter.
	//
	// A map containing the representations for the parameter.
	// The key is the media type and the value describes it.
	// The map MUST only contain one entry.
	Content map[string]Media `json:"content,omitempty"` // TODO: implement

	// Describes how the parameter value will be serialized
	// depending on the type of the parameter value.
	Style string `json:"style,omitempty"`

	// When this is true, parameter values of type array or object
	// generate separate parameters for each value of the array
	// or key-value pair of the map.
	// For other types of parameters this property has no effect.
	Explode *bool `json:"explode,omitempty"`

	Example  json.RawMessage     `json:"example,omitempty"`
	Examples map[string]*Example `json:"examples,omitempty"`
}

// RequestBody describes a single request body.
type RequestBody struct {
	Ref         string `json:"$ref,omitempty"`
	Description string `json:"description,omitempty"`

	// The content of the request body.
	// The key is a media type or media type range and the value describes it.
	// For requests that match multiple keys, only the most specific key is applicable.
	// e.g. text/plain overrides text/*
	Content map[string]Media `json:"content,omitempty"`

	// Determines if the request body is required in the request. Defaults to false.
	Required bool `json:"required,omitempty"`
}

// Responses is a container for the expected responses of an operation.
// The container maps the HTTP response code to the expected response
type Responses map[string]*Response

// Response describes a single response from an API Operation,
// including design-time, static links to operations based on the response.
type Response struct {
	Ref         string                 `json:"$ref,omitempty"`
	Description string                 `json:"description,omitempty"`
	Header      map[string]interface{} `json:"header,omitempty"` // TODO: implement
	Content     map[string]Media       `json:"content,omitempty"`
	Links       map[string]interface{} `json:"links,omitempty"` // TODO: implement
}

// Media provides schema and examples for the media type identified by its key.
type Media struct {
	// The schema defining the content of the request, response, or parameter.
	Schema   *Schema             `json:"schema,omitempty"`
	Example  json.RawMessage     `json:"example,omitempty"`
	Examples map[string]*Example `json:"examples,omitempty"`
}

// Discriminator discriminates types for OneOf, AllOf, AnyOf.
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

// The Schema Object allows the definition of input and output data types.
// These types can be objects, but also primitives and arrays.
type Schema struct {
	Ref         string `json:"$ref,omitempty"` // ref object
	Description string `json:"description,omitempty"`

	// Value MUST be a string. Multiple types via an array are not supported.
	Type string `json:"type,omitempty"`

	// See Data Type Formats for further details (https://swagger.io/specification/#data-type-format).
	// While relying on JSON Schema's defined formats,
	// the OAS offers a few additional predefined formats.
	Format string `json:"format,omitempty"`

	// Property definitions MUST be a Schema Object and not a standard JSON Schema
	// (inline or referenced).
	Properties Properties `json:"properties,omitempty"`

	// Value can be boolean or object. Inline or referenced schema MUST be of a Schema Object
	// and not a standard JSON Schema. Consistent with JSON Schema, additionalProperties defaults to true.
	AdditionalProperties *AdditionalProperties `json:"additionalProperties,omitempty"`

	// The value of this keyword MUST be an array.
	// This array MUST have at least one element.
	// Elements of this array MUST be strings, and MUST be unique.
	Required []string `json:"required,omitempty"`

	// Value MUST be an object and not an array.
	// Inline or referenced schema MUST be of a Schema Object and not a standard JSON Schema.
	// MUST be present if the Type is "array".
	Items *Schema `json:"items,omitempty"`

	// A true value adds "null" to the allowed type specified by the type keyword,
	// only if type is explicitly defined within the same Schema Object.
	// Other Schema Object constraints retain their defined behavior,
	// and therefore may disallow the use of null as a value.
	// A false value leaves the specified or default type unmodified.
	// The default value is false.
	Nullable bool `json:"nullable,omitempty"`

	// AllOf takes an array of object definitions that are used
	// for independent validation but together compose a single object.
	// Still, it does not imply a hierarchy between the models.
	// For that purpose, you should include the discriminator.
	AllOf []*Schema `json:"allOf,omitempty"` // TODO: implement.

	// OneOf validates the value against exactly one of the subschemas
	OneOf []*Schema `json:"oneOf,omitempty"` // TODO: implement.

	// AnyOf validates the value against any (one or more) of the subschemas
	AnyOf []*Schema `json:"anyOf,omitempty"` // TODO: implement.

	// Discriminator for subschemas.
	Discriminator *Discriminator `json:"discriminator,omitempty"`

	// The value of this keyword MUST be an array.
	// This array SHOULD have at least one element.
	// Elements in the array SHOULD be unique.
	Enum []json.RawMessage `json:"enum,omitempty"` // TODO: Nullable.

	// The value of "multipleOf" MUST be a number, strictly greater than 0.
	//
	// A numeric instance is only valid if division by this keyword's value
	// results in an integer.
	MultipleOf *int `json:"multipleOf,omitempty"`

	// The value of "maximum" MUST be a number, representing an upper limit
	// for a numeric instance.
	//
	// If the instance is a number, then this keyword validates if
	// "exclusiveMaximum" is true and instance is less than the provided
	// value, or else if the instance is less than or exactly equal to the
	// provided value.
	Maximum *int64 `json:"maximum,omitempty"`

	// The value of "exclusiveMaximum" MUST be a boolean, representing
	// whether the limit in "maximum" is exclusive or not.  An undefined
	// value is the same as false.
	//
	// If "exclusiveMaximum" is true, then a numeric instance SHOULD NOT be
	// equal to the value specified in "maximum".  If "exclusiveMaximum" is
	// false (or not specified), then a numeric instance MAY be equal to the
	// value of "maximum".
	ExclusiveMaximum bool `json:"exclusiveMaximum,omitempty"`

	// The value of "minimum" MUST be a number, representing a lower limit
	// for a numeric instance.
	//
	// If the instance is a number, then this keyword validates if
	// "exclusiveMinimum" is true and instance is greater than the provided
	// value, or else if the instance is greater than or exactly equal to
	// the provided value.
	Minimum *int64 `json:"minimum,omitempty"`

	// The value of "exclusiveMinimum" MUST be a boolean, representing
	// whether the limit in "minimum" is exclusive or not.  An undefined
	// value is the same as false.
	//
	// If "exclusiveMinimum" is true, then a numeric instance SHOULD NOT be
	// equal to the value specified in "minimum".  If "exclusiveMinimum" is
	// false (or not specified), then a numeric instance MAY be equal to the
	// value of "minimum".
	ExclusiveMinimum bool `json:"exclusiveMinimum,omitempty"`

	// The value of this keyword MUST be a non-negative integer.
	//
	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// A string instance is valid against this keyword if its length is less
	// than, or equal to, the value of this keyword.
	//
	// The length of a string instance is defined as the number of its
	// characters as defined by RFC 7159 [RFC7159].
	MaxLength *uint64 `json:"maxLength,omitempty"`

	// A string instance is valid against this keyword if its length is
	// greater than, or equal to, the value of this keyword.
	//
	// The length of a string instance is defined as the number of its
	// characters as defined by RFC 7159 [RFC7159].
	//
	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// "minLength", if absent, may be considered as being present with
	// integer value 0.
	MinLength *int64 `json:"minLength,omitempty"`

	// The value of this keyword MUST be a string.  This string SHOULD be a
	// valid regular expression, according to the ECMA 262 regular
	// expression dialect.
	//
	// A string instance is considered valid if the regular expression
	// matches the instance successfully. Recall: regular expressions are
	// not implicitly anchored.
	Pattern string `json:"pattern,omitempty"`

	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// An array instance is valid against "maxItems" if its size is less
	// than, or equal to, the value of this keyword.
	MaxItems *uint64 `json:"maxItems,omitempty"`

	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// An array instance is valid against "minItems" if its size is greater
	// than, or equal to, the value of this keyword.
	//
	// If this keyword is not present, it may be considered present with a
	// value of 0.
	MinItems *uint64 `json:"minItems,omitempty"`

	// The value of this keyword MUST be a boolean.
	//
	// If this keyword has boolean value false, the instance validates
	// successfully.  If it has boolean value true, the instance validates
	// successfully if all of its elements are unique.
	//
	// If not present, this keyword may be considered present with boolean
	// value false.
	UniqueItems bool `json:"uniqueItems,omitempty"`

	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// An object instance is valid against "maxProperties" if its number of
	// properties is less than, or equal to, the value of this keyword.
	MaxProperties *uint64 `json:"maxProperties,omitempty"`

	// The value of this keyword MUST be an integer.  This integer MUST be
	// greater than, or equal to, 0.
	//
	// An object instance is valid against "minProperties" if its number of
	// properties is greater than, or equal to, the value of this keyword.
	//
	// If this keyword is not present, it may be considered present with a
	// value of 0.
	MinProperties *uint64 `json:"minProperties,omitempty"`

	// Default value.
	Default json.RawMessage `json:"default,omitempty"` // TODO: support

	// A free-form property to include an example of an instance for this schema.
	// To represent examples that cannot be naturally represented in JSON or YAML,
	// a string value can be used to contain the example with escaping where necessary.
	Example json.RawMessage `json:"example,omitempty"`

	// Specifies that a schema is deprecated and SHOULD be transitioned out
	// of usage.
	Deprecated bool `json:"deprecated,omitempty"`
}

type Property struct {
	Name   string
	Schema *Schema
}

type Properties []Property

func (p Properties) MarshalJSON() ([]byte, error) {
	e := jx.GetEncoder()
	defer jx.PutEncoder(e)

	e.ObjStart()
	for _, prop := range p {
		e.FieldStart(prop.Name)
		b, err := json.Marshal(prop.Schema)
		if err != nil {
			return nil, errors.Wrap(err, "marshal")
		}
		e.Raw(b)
	}
	e.ObjEnd()
	return e.Bytes(), nil
}

func (p *Properties) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return d.Obj(func(d *jx.Decoder, key string) error {
		s := new(Schema)
		b, err := d.Raw()
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, s); err != nil {
			return err
		}

		*p = append(*p, Property{
			Name:   key,
			Schema: s,
		})
		return nil
	})
}

type AdditionalProperties Schema

func (p AdditionalProperties) MarshalJSON() ([]byte, error) {
	return json.Marshal(Schema(p))
}

func (p *AdditionalProperties) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	switch tt := d.Next(); tt {
	case jx.Object:
	case jx.Bool:
		return nil
	default:
		return errors.Errorf("unexpected type %s", tt.String())
	}

	s := Schema{}
	b, err := d.Raw()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = AdditionalProperties(s)
	return nil
}
