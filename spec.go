package ogen

import (
	"encoding/json"

	"github.com/go-faster/jx"
	"github.com/go-faster/yaml"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

type (
	// Num represents JSON number.
	Num = jsonschema.Num
	// Enum is JSON Schema enum validator description.
	Enum = jsonschema.Enum
	// Default is a default value.
	Default = jsonschema.Default
	// ExampleValue is an example value.
	ExampleValue = jsonschema.Example
	// RawValue is a raw JSON value.
	RawValue = jsonschema.RawValue

	// Extensions is a map of OpenAPI extensions.
	//
	// See https://spec.openapis.org/oas/v3.1.0#specification-extensions.
	Extensions = jsonschema.Extensions
	// OpenAPICommon is a common OpenAPI object fields (extensions and locator).
	OpenAPICommon = jsonschema.OpenAPICommon

	// Locator stores location of JSON value.
	Locator = location.Locator
)

// Spec is the root document object of the OpenAPI document.
//
// See https://spec.openapis.org/oas/v3.1.0#openapi-object.
type Spec struct {
	// REQUIRED. This string MUST be the version number of the OpenAPI Specification
	// that the OpenAPI document uses.
	OpenAPI string `json:"openapi" yaml:"openapi"`
	// Added just to detect v2 openAPI specifications and to pretty print version error.
	Swagger string `json:"swagger,omitempty" yaml:"swagger,omitempty"`
	// REQUIRED. Provides metadata about the API.
	//
	// The metadata MAY be used by tooling as required.
	Info Info `json:"info" yaml:"info"`
	// The default value for the `$schema` keyword within Schema Objects contained within this OAS document.
	JSONSchemaDialect string `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`
	// An array of Server Objects, which provide connectivity information to a target server.
	Servers []Server `json:"servers,omitempty" yaml:"servers,omitempty"`
	// The available paths and operations for the API.
	Paths Paths `json:"paths,omitempty" yaml:"paths,omitempty"`
	// The incoming webhooks that MAY be received as part of this API and that
	// the API consumer MAY choose to implement.
	//
	// Closely related to the `callbacks` feature, this section describes requests initiated other
	// than by an API call, for example by an out of band registration.
	//
	// The key name is a unique string to refer to each webhook, while the (optionally referenced)
	// PathItem Object describes a request that may be initiated by the API provider and the expected responses.
	Webhooks map[string]*PathItem `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
	// An element to hold various schemas for the document.
	Components *Components `json:"components,omitempty" yaml:"components,omitempty"`
	// A declaration of which security mechanisms can be used across the API.
	// The list of values includes alternative security requirement objects that can be used.
	//
	// Only one of the security requirement objects need to be satisfied to authorize a request.
	//
	// Individual operations can override this definition.
	Security SecurityRequirements `json:"security,omitempty" yaml:"security,omitempty"`

	// A list of tags used by the specification with additional metadata.
	// The order of the tags can be used to reflect on their order by the parsing
	// tools. Not all tags that are used by the Operation Object must be declared.
	// The tags that are not declared MAY be organized randomly or based on the tools' logic.
	// Each tag name in the list MUST be unique.
	Tags []Tag `json:"tags,omitempty" yaml:"tags,omitempty"`

	// Additional external documentation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`

	// Raw YAML node. Used by '$ref' resolvers.
	Raw *yaml.Node `json:"-" yaml:"-"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (s *Spec) UnmarshalYAML(n *yaml.Node) error {
	type Alias Spec
	var a Alias

	if err := n.Decode(&a); err != nil {
		return err
	}
	*s = Spec(a)
	s.Raw = n

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Spec) UnmarshalJSON(bytes []byte) error {
	type Alias Spec
	var a Alias

	if err := json.Unmarshal(bytes, &a); err != nil {
		return err
	}
	*s = Spec(a)

	var n yaml.Node
	if err := yaml.Unmarshal(bytes, &n); err != nil {
		return err
	}
	s.Raw = &n

	return nil
}

// Init components of schema.
func (s *Spec) Init() {
	if s.Components == nil {
		s.Components = &Components{}
	}
	s.Components.Init()
}

// Info provides metadata about the API.
//
// The metadata MAY be used by the clients if needed,
// and MAY be presented in editing or documentation generation tools for convenience.
//
// See https://spec.openapis.org/oas/v3.1.0#info-object.
type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title" yaml:"title"`
	// A short summary of the API.
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
	// A short description of the API.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// A URL to the Terms of Service for the API. MUST be in the format of a URL.
	TermsOfService string `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	// The contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	// The license information for the exposed API.
	License *License `json:"license,omitempty" yaml:"license,omitempty"`
	// REQUIRED. The version of the OpenAPI document.
	Version string `json:"version" yaml:"version"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`
}

// Contact information for the exposed API.
//
// See https://spec.openapis.org/oas/v3.1.0#contact-object.
type Contact struct {
	// The identifying name of the contact person/organization.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// The URL pointing to the contact information.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	// The email address of the contact person/organization.
	Email string `json:"email,omitempty" yaml:"email,omitempty"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`
}

// License information for the exposed API.
//
// See https://spec.openapis.org/oas/v3.1.0#license-object.
type License struct {
	// REQUIRED. The license name used for the API.
	Name string `json:"name" yaml:"name"`
	// An SPDX license expression for the API.
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	// A URL to the license used for the API.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`
}

// Server represents a Server.
//
// See https://spec.openapis.org/oas/v3.1.0#server-object.
type Server struct {
	// REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative,
	// to indicate that the host location is relative to the location where the OpenAPI document is being served.
	// Variable substitutions will be made when a variable is named in {brackets}.
	URL string `json:"url" yaml:"url"`
	// An optional string describing the host designated by the URL.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// A map between a variable name and its value. The value is used for substitution in the server's URL template.
	Variables map[string]ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// ServerVariable describes an object representing a Server Variable for server URL template substitution.
//
// See https://spec.openapis.org/oas/v3.1.0#server-variable-object
type ServerVariable struct {
	// An enumeration of string values to be used if the substitution options are from a limited set.
	//
	// The array MUST NOT be empty.
	Enum []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	// REQUIRED. The default value to use for substitution, which SHALL be sent if an alternate value is not supplied.
	// Note this behavior is different than the Schema Object's treatment of default values, because in those
	// cases parameter values are optional. If the enum is defined, the value MUST exist in the enum's values.
	Default string `json:"default" yaml:"default"`
	// An optional description for the server variable. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// Components Holds a set of reusable objects for different aspects of the OAS.
// All objects defined within the components object will have no effect on the API unless
// they are explicitly referenced from properties outside the components object.
//
// See https://spec.openapis.org/oas/v3.1.0#components-object.
type Components struct {
	// An object to hold reusable Schema Objects.
	Schemas map[string]*Schema `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	// An object to hold reusable Response Objects.
	Responses map[string]*Response `json:"responses,omitempty" yaml:"responses,omitempty"`
	// An object to hold reusable Parameter Objects.
	Parameters map[string]*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// An object to hold reusable Example Objects.
	Examples map[string]*Example `json:"examples,omitempty" yaml:"examples,omitempty"`
	// An object to hold reusable Request Body Objects.
	RequestBodies map[string]*RequestBody `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	// An object to hold reusable Header Objects.
	Headers map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	// An object to hold reusable Security Scheme Objects.
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	// An object to hold reusable Link Objects.
	Links map[string]*Link `json:"links,omitempty" yaml:"links,omitempty"`
	// An object to hold reusable Callback Objects.
	Callbacks map[string]*Callback `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	// An object to hold reusable Path Item Objects.
	PathItems map[string]*PathItem `json:"pathItems,omitempty" yaml:"pathItems,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

func initMapIfNil[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		m = make(map[K]V)
	}
	return m
}

// Init initializes all fields.
func (c *Components) Init() {
	if c == nil {
		return
	}
	c.Schemas = initMapIfNil(c.Schemas)
	c.Responses = initMapIfNil(c.Responses)
	c.Parameters = initMapIfNil(c.Parameters)
	c.Examples = initMapIfNil(c.Examples)
	c.RequestBodies = initMapIfNil(c.RequestBodies)
	c.Headers = initMapIfNil(c.Headers)
	c.SecuritySchemes = initMapIfNil(c.SecuritySchemes)
	c.Links = initMapIfNil(c.Links)
	c.Callbacks = initMapIfNil(c.Callbacks)
	c.PathItems = initMapIfNil(c.PathItems)
}

// Paths holds the relative paths to the individual endpoints and their operations.
// The path is appended to the URL from the Server Object in order to construct the full URL.
// The Paths MAY be empty, due to ACL constraints.
//
// See https://spec.openapis.org/oas/v3.1.0#paths-object.
type Paths map[string]*PathItem

// PathItem Describes the operations available on a single path.
// A Path Item MAY be empty, due to ACL constraints. The path itself is still exposed to the
// documentation viewer, but they will not know which operations and parameters are available.
//
// See https://spec.openapis.org/oas/v3.1.0#path-item-object.
type PathItem struct {
	// Allows for an external definition of this path item.
	// The referenced structure MUST be in the format of a Path Item Object.
	// In case a Path Item Object field appears both
	// in the defined object and the referenced object, the behavior is undefined.
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	// An optional, string summary, intended to apply to all operations in this path.
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
	// An optional, string description, intended to apply to all operations in this path.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// A definition of a GET operation on this path.
	Get *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	// A definition of a PUT operation on this path.
	Put *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	// A definition of a POST operation on this path.
	Post *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	// A definition of a DELETE operation on this path.
	Delete *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	// A definition of a OPTIONS operation on this path.
	Options *Operation `json:"options,omitempty" yaml:"options,omitempty"`
	// A definition of a HEAD operation on this path.
	Head *Operation `json:"head,omitempty" yaml:"head,omitempty"`
	// A definition of a PATCH operation on this path.
	Patch *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	// A definition of a TRACE operation on this path.
	Trace *Operation `json:"trace,omitempty" yaml:"trace,omitempty"`
	// An alternative server array to service all operations in this path.
	Servers []Server `json:"servers,omitempty" yaml:"servers,omitempty"`
	// A list of parameters that are applicable for all the operations described under this path.
	//
	// These parameters can be overridden at the operation level, but cannot be removed there.
	//
	// The list MUST NOT include duplicated parameters. A unique parameter is defined by
	// a combination of a name and location.
	Parameters []*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// MarshalJSON implements [json.Marshaler].
func (s *PathItem) MarshalJSON() ([]byte, error) {
	type Alias PathItem
	originalJSON, err := json.Marshal(Alias(*s))
	if err != nil {
		return nil, err
	}

	d := jx.DecodeBytes(originalJSON)
	e := jx.Encoder{}

	e.ObjStart()
	if err := d.Obj(func(d *jx.Decoder, key string) error {
		e.FieldStart(key)
		raw, err := d.Raw()
		if err != nil {
			return err
		}

		e.Raw(raw)
		return nil
	}); err != nil {
		return nil, err
	}

	for extK, extV := range s.Common.Extensions {
		e.FieldStart(extK)
		e.Str(extV.Value)
	}

	e.ObjEnd()

	return e.Bytes(), nil
}

// Operation describes a single API operation on a path.
//
// See https://spec.openapis.org/oas/v3.1.0#operation-object.
type Operation struct {
	// A list of tags for API documentation control.
	// Tags can be used for logical grouping of operations by resources or any other qualifier.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	// A short summary of what the operation does.
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
	// A verbose explanation of the operation behavior.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Additional external documentation for this operation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	// Unique string used to identify the operation.
	//
	// The id MUST be unique among all operations described in the API.
	//
	// The operationId value is case-sensitive.
	OperationID string `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	// A list of parameters that are applicable for this operation.
	//
	// If a parameter is already defined at the Path Item, the new definition will override it but
	// can never remove it.
	//
	// The list MUST NOT include duplicated parameters. A unique parameter is defined by
	// a combination of a name and location.
	Parameters []*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// The request body applicable for this operation.
	RequestBody *RequestBody `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	// The list of possible responses as they are returned from executing this operation.
	Responses Responses `json:"responses,omitempty" yaml:"responses,omitempty"`
	// A map of possible out-of band callbacks related to the parent operation.
	//
	// The key is a unique identifier for the Callback Object.
	Callbacks map[string]*Callback `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	// Declares this operation to be deprecated
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	// A declaration of which security mechanisms can be used for this operation.
	//
	// The list of values includes alternative security requirement objects that can be used.
	//
	// Only one of the security requirement objects need to be satisfied to authorize a request.
	Security SecurityRequirements `json:"security,omitempty" yaml:"security,omitempty"`
	// An alternative server array to service this operation.
	//
	// If an alternative server object is specified at the Path Item Object or Root level,
	// it will be overridden by this value.
	Servers []Server `json:"servers,omitempty" yaml:"servers,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// MarshalJSON implements [json.Marshaler].
func (s *Operation) MarshalJSON() ([]byte, error) {
	type Alias Operation
	originalJSON, err := json.Marshal(Alias(*s))
	if err != nil {
		return nil, err
	}

	d := jx.DecodeBytes(originalJSON)
	e := jx.Encoder{}

	e.ObjStart()
	if err := d.Obj(func(d *jx.Decoder, key string) error {
		e.FieldStart(key)
		raw, err := d.Raw()
		if err != nil {
			return err
		}
		e.Raw(raw)
		return nil
	}); err != nil {
		return nil, err
	}

	for extK, extV := range s.Common.Extensions {
		e.FieldStart(extK)
		e.Str(extV.Value)
	}

	e.ObjEnd()

	return e.Bytes(), nil
}

// ExternalDocumentation describes a reference to external resource for extended documentation.
//
// See https://spec.openapis.org/oas/v3.1.0#external-documentation-object.
type ExternalDocumentation struct {
	// A description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// REQUIRED. The URL for the target documentation. This MUST be in the form of a URL.
	URL string `json:"url" yaml:"url"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`
}

// Parameter describes a single operation parameter.
// A unique parameter is defined by a combination of a name and location.
//
// See https://spec.openapis.org/oas/v3.1.0#parameter-object.
type Parameter struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"` // ref object

	// REQUIRED. The name of the parameter. Parameter names are case sensitive.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In string `json:"in,omitempty" yaml:"in,omitempty"`
	// A brief description of the parameter. This could contain examples of use.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Determines whether this parameter is mandatory.
	// If the parameter location is "path", this property is REQUIRED
	// and its value MUST be true.
	// Otherwise, the property MAY be included and its default value is false.
	Required bool `json:"required,omitempty" yaml:"required,omitempty"`
	// Specifies that a parameter is deprecated and SHOULD be transitioned out of usage.
	// Default value is false.
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`

	// Describes how the parameter value will be serialized
	// depending on the type of the parameter value.
	Style string `json:"style,omitempty" yaml:"style,omitempty"`
	// When this is true, parameter values of type array or object
	// generate separate parameters for each value of the array
	// or key-value pair of the map.
	// For other types of parameters this property has no effect.
	Explode *bool `json:"explode,omitempty" yaml:"explode,omitempty"`
	// Determines whether the parameter value SHOULD allow reserved characters, as defined by RFC 3986.
	//
	// This property only applies to parameters with an in value of query.
	//
	// The default value is false.
	AllowReserved bool `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	// The schema defining the type used for the parameter.
	Schema *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
	// Example of the parameter's potential value.
	Example ExampleValue `json:"example,omitempty" yaml:"example,omitempty"`
	// Examples of the parameter's potential value.
	Examples map[string]*Example `json:"examples,omitempty" yaml:"examples,omitempty"`

	// For more complex scenarios, the content property can define the media type and schema of the parameter.
	// A parameter MUST contain either a schema property, or a content property, but not both.
	// When example or examples are provided in conjunction with the schema object,
	// the example MUST follow the prescribed serialization strategy for the parameter.
	//
	// A map containing the representations for the parameter.
	// The key is the media type and the value describes it.
	// The map MUST only contain one entry.
	Content map[string]Media `json:"content,omitempty" yaml:"content,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// RequestBody describes a single request body.
//
// See https://spec.openapis.org/oas/v3.1.0#request-body-object.
type RequestBody struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"` // ref object

	// A brief description of the request body. This could contain examples of use.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// REQUIRED. The content of the request body.
	//
	// The key is a media type or media type range and the value describes it.
	//
	// For requests that match multiple keys, only the most specific key is applicable.
	// e.g. text/plain overrides text/*
	Content map[string]Media `json:"content,omitempty" yaml:"content,omitempty"`

	// Determines if the request body is required in the request. Defaults to false.
	Required bool `json:"required,omitempty" yaml:"required,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// Media provides schema and examples for the media type identified by its key.
//
// See https://spec.openapis.org/oas/v3.1.0#media-type-object.
type Media struct {
	// The schema defining the content of the request, response, or parameter.
	Schema *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
	// Example of the media type.
	Example ExampleValue `json:"example,omitempty" yaml:"example,omitempty"`
	// Examples of the media type.
	Examples map[string]*Example `json:"examples,omitempty" yaml:"examples,omitempty"`

	// A map between a property name and its encoding information.
	//
	// The key, being the property name, MUST exist in the schema as a property.
	//
	// The encoding object SHALL only apply to requestBody objects when the media
	// type is multipart or application/x-www-form-urlencoded.
	Encoding map[string]Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// Encoding describes single encoding definition applied to a single schema property.
//
// See https://spec.openapis.org/oas/v3.1.0#encoding-object.
type Encoding struct {
	// The Content-Type for encoding a specific property.
	ContentType string `json:"contentType,omitempty" yaml:"contentType,omitempty"`

	// A map allowing additional information to be provided as headers, for example Content-Disposition.
	// Content-Type is described separately and SHALL be ignored in this section. This property SHALL be
	// ignored if the request body media type is not a multipart.
	Headers map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Describes how the parameter value will be serialized
	// depending on the type of the parameter value.
	Style string `json:"style,omitempty" yaml:"style,omitempty"`

	// When this is true, parameter values of type array or object
	// generate separate parameters for each value of the array
	// or key-value pair of the map.
	// For other types of parameters this property has no effect.
	Explode *bool `json:"explode,omitempty" yaml:"explode,omitempty"`

	// Determines whether the parameter value SHOULD allow reserved characters, as defined by
	// RFC3986 :/?#[]@!$&'()*+,;= to be included without percent-encoding.
	// The default value is false. This property SHALL be ignored if the request body media type
	// is not application/x-www-form-urlencoded.
	AllowReserved bool `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// Responses is a container for the expected responses of an operation.
//
// The container maps the HTTP response code to the expected response.
//
// The `default` MAY be used as a default response object for all HTTP
// codes that are not covered individually by the Responses Object.
//
// The Responses Object MUST contain at least one response code, and if only one
// response code is provided it SHOULD be the response for a successful operation call.
//
// See https://spec.openapis.org/oas/v3.1.0#responses-object.
type Responses map[string]*Response

// Response describes a single response from an API Operation,
// including design-time, static links to operations based on the response.
//
// See https://spec.openapis.org/oas/v3.1.0#response-object.
type Response struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"` // ref object

	// REQUIRED. A description of the response.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Maps a header name to its definition.
	//
	// RFC7230 states header names are case insensitive.
	//
	// If a response header is defined with the name "Content-Type", it SHALL be ignored.
	Headers map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	// A map containing descriptions of potential response payloads.
	//
	// The key is a media type or media type range and the value describes it.
	//
	// For requests that match multiple keys, only the most specific key is applicable.
	// e.g. text/plain overrides text/*
	Content map[string]Media `json:"content,omitempty" yaml:"content,omitempty"`
	// A map of operations links that can be followed from the response.
	//
	// The key of the map is a short name for the link, following the naming constraints
	// of the names for Component Objects.
	Links map[string]*Link `json:"links,omitempty" yaml:"links,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// Callback is a map of possible out-of band callbacks related to the parent operation.
//
// Each value in the map is a Path Item Object that describes a set of requests that may be
// initiated by the API provider and the expected responses.
//
// The key value used to identify the path item object is an expression, evaluated at runtime,
// that identifies a URL to use for the callback operation.
//
// To describe incoming requests from the API provider independent from another
// API call, use the `webhooks` field.
//
// See https://spec.openapis.org/oas/v3.1.0#callback-object.
type Callback map[string]*PathItem

// Example object.
//
// See https://spec.openapis.org/oas/v3.1.0#example-object.
type Example struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"` // ref object

	// Short description for the example.
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
	// Long description for the example.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Embedded literal example.
	Value ExampleValue `json:"value,omitempty" yaml:"value,omitempty"`
	// A URI that points to the literal example.
	ExternalValue string `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// Link describes a possible design-time link for a response.
//
// See https://spec.openapis.org/oas/v3.1.0#link-object.
type Link struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"` // ref object

	// A relative or absolute URI reference to an OAS operation.
	//
	// This field is mutually exclusive of the operationId field, and MUST point to an Operation Object.
	OperationRef string `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	// The name of an existing, resolvable OAS operation, as defined with a unique operationId.
	//
	// This field is mutually exclusive of the operationRef field.
	OperationID string `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	// A map representing parameters to pass to an operation as specified with operationId or identified
	// via operationRef.
	//
	// The key is the parameter name to be used, whereas the value can be a constant or an expression to be
	// evaluated and passed to the linked operation.
	Parameters map[string]RawValue `json:"parameters,omitempty" yaml:"parameters,omitempty"`

	Common OpenAPICommon `json:"-" yaml:",inline"`
}

// Header describes header response.
//
// Header Object follows the structure of the Parameter Object with the following changes:
//
//  1. `name` MUST NOT be specified, it is given in the corresponding headers map.
//  2. `in` MUST NOT be specified, it is implicitly in header.
//  3. All traits that are affected by the location MUST be applicable to a location of header.
//
// See https://spec.openapis.org/oas/v3.1.0#header-object.
type Header = Parameter

// Tag adds metadata to a single tag that is used by the Operation Object.
//
// See https://spec.openapis.org/oas/v3.1.0#tag-object
type Tag struct {
	// REQUIRED. The name of the tag.
	Name string `json:"name" yaml:"name"`
	// A description for the tag. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Additional external documentation for this tag.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	// Specification extensions.
	Extensions Extensions `json:"-" yaml:",inline"`
}
