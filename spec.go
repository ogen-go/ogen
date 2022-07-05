package ogen

import (
	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"

	ogenjson "github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/jsonschema"
)

type (
	// Num represents JSON number.
	Num = jsonschema.Num
	// Enum is JSON Schema enum validator description.
	Enum = jsonschema.Enum
	// Locator stores location of JSON value.
	Locator = ogenjson.Locator
)

// Spec is the root document object of the OpenAPI document.
type Spec struct {
	// This string MUST be the semantic version number
	// of the OpenAPI Specification version that the OpenAPI document uses.
	OpenAPI    string               `json:"openapi"`
	Info       Info                 `json:"info"`
	Servers    []Server             `json:"servers,omitzero"`
	Paths      Paths                `json:"paths,omitzero"`
	Components *Components          `json:"components,omitzero"`
	Security   SecurityRequirements `json:"security,omitzero"`

	// A list of tags used by the specification with additional metadata.
	// The order of the tags can be used to reflect on their order by the parsing
	// tools. Not all tags that are used by the Operation Object must be declared.
	// The tags that are not declared MAY be organized randomly or based on the tools' logic.
	// Each tag name in the list MUST be unique.
	Tags []Tag `json:"tags,omitzero"`

	// Additional external documentation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitzero"`

	// Raw JSON value. Used by JSON Schema resolver.
	Raw json.RawValue `json:"-"`
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

	a.Raw = append(a.Raw[:0], value...)
	*s = Spec(a)
	return nil
}

// Init components of schema.
func (s *Spec) Init() {
	if s.Components == nil {
		s.Components = &Components{}
	}
	s.Components.Init()
}

// Example object.
//
// https://swagger.io/specification/#example-object
type Example struct {
	Ref           string        `json:"$ref,omitzero"` // ref object
	Summary       string        `json:"summary,omitzero"`
	Description   string        `json:"description,omitzero"`
	Value         json.RawValue `json:"value,omitzero"`
	ExternalValue string        `json:"externalValue,omitzero"`

	Locator `json:"-"`
}

// Tag object.
//
// https://swagger.io/specification/#tag-object
type Tag struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitzero"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitzero"`
}

// Info provides metadata about the API.
//
// The metadata MAY be used by the clients if needed,
// and MAY be presented in editing or documentation generation tools for convenience.
type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title"`
	// A short summary of the API.
	Summary string `json:"summary,omitzero"`
	// A short description of the API.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitzero"`
	// A URL to the Terms of Service for the API. MUST be in the format of a URL.
	TermsOfService string `json:"termsOfService,omitzero"`
	// The contact information for the exposed API.
	Contact *Contact `json:"contact,omitzero"`
	// The license information for the exposed API.
	License *License `json:"license,omitzero"`
	// REQUIRED. The version of the OpenAPI document.
	Version string `json:"version"`
}

// Contact information for the exposed API.
type Contact struct {
	Name  string `json:"name,omitzero"`
	URL   string `json:"url,omitzero"`
	Email string `json:"email,omitzero"`
}

// License information for the exposed API.
type License struct {
	Name string `json:"name,omitzero"`
	URL  string `json:"url,omitzero"`
}

// Server represents a Server.
type Server struct {
	// REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative,
	// to indicate that the host location is relative to the location where the OpenAPI document is being served.
	// Variable substitutions will be made when a variable is named in {brackets}.
	URL string `json:"url"`
	// An optional string describing the host designated by the URL.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitzero"`
	// A map between a variable name and its value. The value is used for substitution in the server's URL template.
	Variables map[string]ServerVariable `json:"variables,omitzero"`
}

// ServerVariable describes an object representing a Server Variable for server URL template substitution.
type ServerVariable struct {
	// An enumeration of string values to be used if the substitution options are from a limited set.
	//
	// The array MUST NOT be empty.
	Enum []string `json:"enum,omitzero"`
	// REQUIRED. The default value to use for substitution, which SHALL be sent if an alternate value is not supplied.
	// Note this behavior is different than the Schema Object’s treatment of default values, because in those
	// cases parameter values are optional. If the enum is defined, the value MUST exist in the enum’s values.
	Default string `json:"default"`
	// An optional description for the server variable. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitzero"`
}

// ExternalDocumentation describes a reference to external resource for extended documentation.
type ExternalDocumentation struct {
	// A description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitzero"`
	// REQUIRED. The URL for the target documentation. This MUST be in the form of a URL.
	URL string `json:"url"`
}

// Components hold a set of reusable objects for different aspects of the OAS.
// All objects defined within the components object will have no effect on the API
// unless they are explicitly referenced from properties outside the components object.
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitzero"`
	Responses       map[string]*Response       `json:"responses,omitzero"`
	Parameters      map[string]*Parameter      `json:"parameters,omitzero"`
	Examples        map[string]*Example        `json:"examples,omitzero"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitzero"`
	Headers         map[string]*Header         `json:"headers,omitzero"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitzero"`
	// Links           map[string]Link            `json:"links"`
	// Callbacks       map[string]Callback        `json:"callback"`
}

// Init initializes all fields.
func (c *Components) Init() {
	if c == nil {
		return
	}
	if c.Schemas == nil {
		c.Schemas = map[string]*Schema{}
	}
	if c.Responses == nil {
		c.Responses = map[string]*Response{}
	}
	if c.Parameters == nil {
		c.Parameters = map[string]*Parameter{}
	}
	if c.Headers == nil {
		c.Headers = map[string]*Header{}
	}
	if c.Examples == nil {
		c.Examples = map[string]*Example{}
	}
	if c.RequestBodies == nil {
		c.RequestBodies = map[string]*RequestBody{}
	}
	if c.SecuritySchemes == nil {
		c.SecuritySchemes = map[string]*SecurityScheme{}
	}
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
	Ref         string       `json:"$ref,omitzero"`
	Summary     string       `json:"summary,omitzero"`
	Description string       `json:"description,omitzero"`
	Get         *Operation   `json:"get,omitzero"`
	Put         *Operation   `json:"put,omitzero"`
	Post        *Operation   `json:"post,omitzero"`
	Delete      *Operation   `json:"delete,omitzero"`
	Options     *Operation   `json:"options,omitzero"`
	Head        *Operation   `json:"head,omitzero"`
	Patch       *Operation   `json:"patch,omitzero"`
	Trace       *Operation   `json:"trace,omitzero"`
	Servers     []Server     `json:"servers,omitzero"`
	Parameters  []*Parameter `json:"parameters,omitzero"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	OperationID string               `json:"operationId,omitzero"`
	Security    SecurityRequirements `json:"security,omitzero"`
	Parameters  []*Parameter         `json:"parameters,omitzero"`
	RequestBody *RequestBody         `json:"requestBody,omitzero"`
	Responses   Responses            `json:"responses,omitzero"`

	// A list of tags for API documentation control.
	// Tags can be used for logical grouping of operations by resources or any other qualifier.
	Tags         []string               `json:"tags,omitzero"`
	Summary      string                 `json:"summary,omitzero"`
	Description  string                 `json:"description,omitzero"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitzero"`
	Deprecated   bool                   `json:"deprecated,omitzero"`

	Locator `json:"-"`
}

// Parameter describes a single operation parameter.
// A unique parameter is defined by a combination of a name and location.
type Parameter struct {
	Ref  string `json:"$ref,omitzero"`
	Name string `json:"name"`

	// The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In          string  `json:"in"`
	Description string  `json:"description,omitzero"`
	Schema      *Schema `json:"schema,omitzero"`

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
	Content map[string]Media `json:"content,omitzero"`

	// Describes how the parameter value will be serialized
	// depending on the type of the parameter value.
	Style string `json:"style,omitzero"`

	// When this is true, parameter values of type array or object
	// generate separate parameters for each value of the array
	// or key-value pair of the map.
	// For other types of parameters this property has no effect.
	Explode *bool `json:"explode,omitzero"`

	Example  json.RawValue       `json:"example,omitzero"`
	Examples map[string]*Example `json:"examples,omitzero"`

	Locator `json:"-"`
}

// RequestBody describes a single request body.
type RequestBody struct {
	Ref         string `json:"$ref,omitzero"`
	Description string `json:"description,omitzero"`

	// The content of the request body.
	// The key is a media type or media type range and the value describes it.
	// For requests that match multiple keys, only the most specific key is applicable.
	// e.g. text/plain overrides text/*
	Content map[string]Media `json:"content,omitzero"`

	// Determines if the request body is required in the request. Defaults to false.
	Required bool `json:"required,omitzero"`

	Locator `json:"-"`
}

// Responses is a container for the expected responses of an operation.
// The container maps the HTTP response code to the expected response
type Responses map[string]*Response

// Response describes a single response from an API Operation,
// including design-time, static links to operations based on the response.
type Response struct {
	Ref         string                 `json:"$ref,omitzero"`
	Description string                 `json:"description,omitzero"`
	Headers     map[string]*Header     `json:"headers,omitzero"`
	Content     map[string]Media       `json:"content,omitzero"`
	Links       map[string]interface{} `json:"links,omitzero"` // TODO: implement

	Locator `json:"-"`
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
	Schema   *Schema             `json:"schema,omitzero"`
	Example  json.RawValue       `json:"example,omitzero"`
	Examples map[string]*Example `json:"examples,omitzero"`

	// A map between a property name and its encoding information. The key, being the property name, MUST exist in
	// the schema as a property. The encoding object SHALL only apply to requestBody objects when the media
	// type is multipart or application/x-www-form-urlencoded.
	Encoding map[string]Encoding `json:"encoding,omitzero"`

	Locator `json:"-"`
}

// Encoding describes single encoding definition applied to a single schema property.
type Encoding struct {
	// The Content-Type for encoding a specific property.
	ContentType string `json:"contentType,omitzero"`

	// A map allowing additional information to be provided as headers, for example Content-Disposition.
	// Content-Type is described separately and SHALL be ignored in this section. This property SHALL be
	// ignored if the request body media type is not a multipart.
	Headers map[string]*Header `json:"headers,omitzero"`

	// Describes how the parameter value will be serialized
	// depending on the type of the parameter value.
	Style string `json:"style,omitzero"`

	// When this is true, parameter values of type array or object
	// generate separate parameters for each value of the array
	// or key-value pair of the map.
	// For other types of parameters this property has no effect.
	Explode *bool `json:"explode,omitzero"`

	// Determines whether the parameter value SHOULD allow reserved characters, as defined by
	// RFC3986 :/?#[]@!$&'()*+,;= to be included without percent-encoding.
	// The default value is false. This property SHALL be ignored if the request body media type
	// is not application/x-www-form-urlencoded.
	AllowReserved bool `json:"allowReserved,omitzero"`

	Locator `json:"-"`
}

// Discriminator discriminates types for OneOf, AllOf, AnyOf.
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitzero"`
}
