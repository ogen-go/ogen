package ogen

import (
	"reflect"

	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"

	"github.com/ogen-go/ogen/jsonschema"
)

type (
	// Num represents JSON number.
	Num = jsonschema.Num
	// Enum is JSON Schema enum validator description.
	Enum = jsonschema.Enum
)

// Spec is the root document object of the OpenAPI document.
type Spec struct {
	// This string MUST be the semantic version number
	// of the OpenAPI Specification version that the OpenAPI document uses.
	OpenAPI    string               `json:"openapi"`
	Info       Info                 `json:"info"`
	Servers    []Server             `json:"servers,omitempty"`
	Paths      Paths                `json:"paths,omitempty"`
	Components *Components          `json:"components,omitempty"`
	Security   SecurityRequirements `json:"security,omitempty"`

	// A list of tags used by the specification with additional metadata.
	// The order of the tags can be used to reflect on their order by the parsing
	// tools. Not all tags that are used by the Operation Object must be declared.
	// The tags that are not declared MAY be organized randomly or based on the tools' logic.
	// Each tag name in the list MUST be unique.
	Tags []Tag `json:"tags,omitempty"`

	// Additional external documentation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`

	// Raw JSON value. Used by JSON Schema resolver.
	Raw []byte `json:"-"`
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (s *Spec) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	type Alias Spec
	var a Alias

	value, err := d.ReadValue()
	if err != nil {
		return err
	}
	if err := opts.Unmarshal(json.DecodeOptions{}, value, &a); err != nil {
		return errors.Wrap(err, "spec")
	}

	a.Raw = append(a.Raw, value...)
	*s = Spec(a)
	return nil
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
	Ref           string        `json:"$ref,omitempty"` // ref object
	Summary       string        `json:"summary,omitempty"`
	Description   string        `json:"description,omitempty"`
	Value         json.RawValue `json:"value,omitempty"`
	ExternalValue string        `json:"externalValue,omitempty"`
}

// Tag object.
//
// https://swagger.io/specification/#tag-object
type Tag struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// Info provides metadata about the API.
//
// The metadata MAY be used by the clients if needed,
// and MAY be presented in editing or documentation generation tools for convenience.
type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title"`
	// A short summary of the API.
	Summary string `json:"summary,omitempty"`
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
	// REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative,
	// to indicate that the host location is relative to the location where the OpenAPI document is being served.
	// Variable substitutions will be made when a variable is named in {brackets}.
	URL string `json:"url"`
	// An optional string describing the host designated by the URL.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// A map between a variable name and its value. The value is used for substitution in the server's URL template.
	Variables map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable describes an object representing a Server Variable for server URL template substitution.
type ServerVariable struct {
	// An enumeration of string values to be used if the substitution options are from a limited set.
	//
	// The array MUST NOT be empty.
	Enum []string `json:"enum,omitempty"`
	// REQUIRED. The default value to use for substitution, which SHALL be sent if an alternate value is not supplied.
	// Note this behavior is different than the Schema Object’s treatment of default values, because in those
	// cases parameter values are optional. If the enum is defined, the value MUST exist in the enum’s values.
	Default string `json:"default"`
	// An optional description for the server variable. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
}

// Components hold a set of reusable objects for different aspects of the OAS.
// All objects defined within the components object will have no effect on the API
// unless they are explicitly referenced from properties outside the components object.
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty"`
	Headers         map[string]*Header         `json:"headers,omitempty"`
	Examples        map[string]*Example        `json:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
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
	Summary     string       `json:"summary,omitempty"`
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
	Tags []string `json:"tags,omitempty"`

	Summary      string                 `json:"summary,omitempty"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`

	OperationID string               `json:"operationId,omitempty"`
	Parameters  []*Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody         `json:"requestBody,omitempty"`
	Responses   Responses            `json:"responses,omitempty"`
	Security    SecurityRequirements `json:"security,omitempty"`
	Deprecated  bool                 `json:"deprecated,omitempty"`
}

// ExternalDocumentation describes a reference to external resource for extended documentation.
type ExternalDocumentation struct {
	// A description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// REQUIRED. The URL for the target documentation. This MUST be in the form of a URL.
	URL string `json:"url"`
}

// Parameter describes a single operation parameter.
// A unique parameter is defined by a combination of a name and location.
type Parameter struct {
	Ref  string `json:"$ref,omitempty"`
	Name string `json:"name"`

	// The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In          string  `json:"in"`
	Description string  `json:"description,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`

	// Determines whether this parameter is mandatory.
	// If the parameter location is "path", this property is REQUIRED
	// and its value MUST be true.
	// Otherwise, the property MAY be included and its default value is false.
	Required bool `json:"required,omitzero"`

	// Specifies that a parameter is deprecated and SHOULD be transitioned out of usage.
	// Default value is false.
	Deprecated bool `json:"deprecated,omitzero"`

	// For more complex scenarios, the content property can define the media type and schema of the parameter.
	// A parameter MUST contain either a schema property, or a content property, but not both.
	// When example or examples are provided in conjunction with the schema object,
	// the example MUST follow the prescribed serialization strategy for the parameter.
	//
	// A map containing the representations for the parameter.
	// The key is the media type and the value describes it.
	// The map MUST only contain one entry.
	Content map[string]Media `json:"content,omitempty"`

	// Describes how the parameter value will be serialized
	// depending on the type of the parameter value.
	Style string `json:"style,omitempty"`

	// When this is true, parameter values of type array or object
	// generate separate parameters for each value of the array
	// or key-value pair of the map.
	// For other types of parameters this property has no effect.
	Explode *bool `json:"explode,omitempty"`

	Example  json.RawValue       `json:"example,omitempty"`
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
	Required bool `json:"required,omitzero"`
}

// Responses is a container for the expected responses of an operation.
// The container maps the HTTP response code to the expected response
type Responses map[string]*Response

// Response describes a single response from an API Operation,
// including design-time, static links to operations based on the response.
type Response struct {
	Ref         string                 `json:"$ref,omitempty"`
	Description string                 `json:"description,omitempty"`
	Headers     map[string]*Header     `json:"headers,omitempty"`
	Content     map[string]Media       `json:"content,omitempty"`
	Links       map[string]interface{} `json:"links,omitempty"` // TODO: implement
}

// Header describes header response.
//
// Header Object follows the structure of the Parameter Object with the following changes:
//
// 	1. `name` MUST NOT be specified, it is given in the corresponding headers map.
// 	2. `in` MUST NOT be specified, it is implicitly in header.
// 	3. All traits that are affected by the location MUST be applicable to a location of header.
//
type Header = Parameter

// Media provides schema and examples for the media type identified by its key.
type Media struct {
	// The schema defining the content of the request, response, or parameter.
	Schema   *Schema             `json:"schema,omitempty"`
	Example  json.RawValue       `json:"example,omitempty"`
	Examples map[string]*Example `json:"examples,omitempty"`

	// A map between a property name and its encoding information. The key, being the property name, MUST exist in
	// the schema as a property. The encoding object SHALL only apply to requestBody objects when the media
	// type is multipart or application/x-www-form-urlencoded.
	Encoding map[string]Encoding `json:"encoding,omitempty"`
}

// Encoding describes single encoding definition applied to a single schema property.
type Encoding struct {
	// The Content-Type for encoding a specific property.
	ContentType string `json:"contentType,omitempty"`

	// A map allowing additional information to be provided as headers, for example Content-Disposition.
	// Content-Type is described separately and SHALL be ignored in this section. This property SHALL be
	// ignored if the request body media type is not a multipart.
	Headers map[string]*Header `json:"headers,omitempty"`

	// Describes how the parameter value will be serialized
	// depending on the type of the parameter value.
	Style string `json:"style,omitempty"`

	// When this is true, parameter values of type array or object
	// generate separate parameters for each value of the array
	// or key-value pair of the map.
	// For other types of parameters this property has no effect.
	Explode *bool `json:"explode,omitempty"`

	// Determines whether the parameter value SHOULD allow reserved characters, as defined by
	// RFC3986 :/?#[]@!$&'()*+,;= to be included without percent-encoding.
	// The default value is false. This property SHALL be ignored if the request body media type
	// is not application/x-www-form-urlencoded.
	AllowReserved bool `json:"allowReserved,omitzero"`
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
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`

	// Additional external documentation for this schema.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`

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

	// The value of "patternProperties" MUST be an object. Each property
	// name of this object SHOULD be a valid regular expression, according
	// to the ECMA-262 regular expression dialect. Each property value of
	// this object MUST be a valid JSON Schema.
	PatternProperties PatternProperties `json:"patternProperties,omitempty"`

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
	Nullable bool `json:"nullable,omitzero"`

	// AllOf takes an array of object definitions that are used
	// for independent validation but together compose a single object.
	// Still, it does not imply a hierarchy between the models.
	// For that purpose, you should include the discriminator.
	AllOf []*Schema `json:"allOf,omitempty"` // TODO: implement.

	// OneOf validates the value against exactly one of the subschemas
	OneOf []*Schema `json:"oneOf,omitempty"`

	// AnyOf validates the value against any (one or more) of the subschemas
	AnyOf []*Schema `json:"anyOf,omitempty"`

	// Discriminator for subschemas.
	Discriminator *Discriminator `json:"discriminator,omitempty"`

	// The value of this keyword MUST be an array.
	// This array SHOULD have at least one element.
	// Elements in the array SHOULD be unique.
	Enum Enum `json:"enum,omitempty"`

	// The value of "multipleOf" MUST be a number, strictly greater than 0.
	//
	// A numeric instance is only valid if division by this keyword's value
	// results in an integer.
	MultipleOf Num `json:"multipleOf,omitempty"`

	// The value of "maximum" MUST be a number, representing an upper limit
	// for a numeric instance.
	//
	// If the instance is a number, then this keyword validates if
	// "exclusiveMaximum" is true and instance is less than the provided
	// value, or else if the instance is less than or exactly equal to the
	// provided value.
	Maximum Num `json:"maximum,omitempty"`

	// The value of "exclusiveMaximum" MUST be a boolean, representing
	// whether the limit in "maximum" is exclusive or not.  An undefined
	// value is the same as false.
	//
	// If "exclusiveMaximum" is true, then a numeric instance SHOULD NOT be
	// equal to the value specified in "maximum".  If "exclusiveMaximum" is
	// false (or not specified), then a numeric instance MAY be equal to the
	// value of "maximum".
	ExclusiveMaximum bool `json:"exclusiveMaximum,omitzero"`

	// The value of "minimum" MUST be a number, representing a lower limit
	// for a numeric instance.
	//
	// If the instance is a number, then this keyword validates if
	// "exclusiveMinimum" is true and instance is greater than the provided
	// value, or else if the instance is greater than or exactly equal to
	// the provided value.
	Minimum Num `json:"minimum,omitempty"`

	// The value of "exclusiveMinimum" MUST be a boolean, representing
	// whether the limit in "minimum" is exclusive or not.  An undefined
	// value is the same as false.
	//
	// If "exclusiveMinimum" is true, then a numeric instance SHOULD NOT be
	// equal to the value specified in "minimum".  If "exclusiveMinimum" is
	// false (or not specified), then a numeric instance MAY be equal to the
	// value of "minimum".
	ExclusiveMinimum bool `json:"exclusiveMinimum,omitzero"`

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
	MinLength *uint64 `json:"minLength,omitempty"`

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
	UniqueItems bool `json:"uniqueItems,omitzero"`

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
	Default json.RawValue `json:"default,omitempty"`

	// A free-form property to include an example of an instance for this schema.
	// To represent examples that cannot be naturally represented in JSON or YAML,
	// a string value can be used to contain the example with escaping where necessary.
	Example json.RawValue `json:"example,omitempty"`

	// Specifies that a schema is deprecated and SHOULD be transitioned out
	// of usage.
	Deprecated bool `json:"deprecated,omitzero"`

	// If the instance value is a string, this property defines that the
	// string SHOULD be interpreted as binary data and decoded using the
	// encoding named by this property.  RFC 2045, Section 6.1 lists
	// the possible values for this property.
	//
	// The value of this property MUST be a string.
	//
	// The value of this property SHOULD be ignored if the instance
	// described is not a string.
	ContentEncoding string `json:"contentEncoding,omitempty"`

	// The value of this property must be a media type, as defined by RFC
	// 2046. This property defines the media type of instances
	// which this schema defines.
	//
	// The value of this property MUST be a string.
	//
	// The value of this property SHOULD be ignored if the instance
	// described is not a string.
	ContentMediaType string `json:"contentMediaType,omitempty"`
}

// Property is item of Properties.
type Property struct {
	Name   string
	Schema *Schema
}

// Properties represent JSON Schema properties validator description.
type Properties []Property

// MarshalNextJSON implements json.MarshalerV2.
func (p Properties) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	if err := e.WriteToken(json.ObjectStart); err != nil {
		return err
	}
	for _, member := range p {
		if err := opts.MarshalNext(e, member.Name); err != nil {
			return err
		}
		if err := opts.MarshalNext(e, member.Schema); err != nil {
			return err
		}
	}
	if err := e.WriteToken(json.ObjectEnd); err != nil {
		return err
	}
	return nil
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (p *Properties) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) (rerr error) {
	offset := d.InputOffset()
	if kind := d.PeekKind(); kind != '{' {
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: d.StackPointer(),
			JSONKind:    kind,
			GoType:      reflect.TypeOf(p),
			Err:         errors.Errorf("unexpected type %s", kind.String()),
		}
	}

	// Read the opening brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	// Keep non-nil value, to distinguish from not set object.
	properties := Properties{}
	for d.PeekKind() != '}' {
		var (
			name   string
			schema *Schema
		)
		if err := opts.UnmarshalNext(d, &name); err != nil {
			return err
		}
		if err := opts.UnmarshalNext(d, &schema); err != nil {
			return err
		}
		properties = append(properties, Property{Name: name, Schema: schema})
	}
	// Read the closing brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	*p = properties
	return nil
}

// AdditionalProperties represent JSON Schema additionalProperties validator description.
type AdditionalProperties struct {
	Bool   *bool
	Schema Schema
}

// MarshalNextJSON implements json.MarshalerV2.
func (p AdditionalProperties) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	if p.Bool != nil {
		return opts.MarshalNext(e, p.Bool)
	}
	return opts.MarshalNext(e, p.Schema)
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (p *AdditionalProperties) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	offset := d.InputOffset()
	switch kind := d.PeekKind(); kind {
	case 't', 'f':
		return opts.UnmarshalNext(d, &p.Bool)
	case '{':
		return opts.UnmarshalNext(d, &p.Schema)
	default:
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: d.StackPointer(),
			JSONKind:    kind,
			GoType:      reflect.TypeOf(p),
			Err:         errors.Errorf("unexpected type %s", kind.String()),
		}
	}
}

// PatternProperty is item of PatternProperties.
type PatternProperty struct {
	Pattern string
	Schema  *Schema
}

// PatternProperties represent JSON Schema patternProperties validator description.
type PatternProperties []PatternProperty

// MarshalNextJSON implements json.MarshalerV2.
func (r PatternProperties) MarshalNextJSON(opts json.MarshalOptions, e *json.Encoder) error {
	if err := e.WriteToken(json.ObjectStart); err != nil {
		return err
	}
	for _, member := range r {
		if err := opts.MarshalNext(e, member.Pattern); err != nil {
			return err
		}
		if err := opts.MarshalNext(e, member.Schema); err != nil {
			return err
		}
	}
	if err := e.WriteToken(json.ObjectEnd); err != nil {
		return err
	}
	return nil
}

// UnmarshalNextJSON implements json.UnmarshalerV2.
func (r *PatternProperties) UnmarshalNextJSON(opts json.UnmarshalOptions, d *json.Decoder) error {
	offset := d.InputOffset()
	if kind := d.PeekKind(); kind != '{' {
		return &json.SemanticError{
			ByteOffset:  offset,
			JSONPointer: d.StackPointer(),
			JSONKind:    kind,
			GoType:      reflect.TypeOf(r),
			Err:         errors.Errorf("unexpected type %s", kind.String()),
		}
	}
	// Read the opening brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	// Keep non-nil value, to distinguish from not set object.
	patternProperties := PatternProperties{}
	for d.PeekKind() != '}' {
		var (
			pattern string
			schema  *Schema
		)
		if err := opts.UnmarshalNext(d, &pattern); err != nil {
			return err
		}
		if err := opts.UnmarshalNext(d, &schema); err != nil {
			return err
		}
		patternProperties = append(patternProperties, PatternProperty{Pattern: pattern, Schema: schema})
	}
	// Read the closing brace.
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	return nil
}
