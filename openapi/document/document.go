// Package document contains the raw types of OpenAPI v3 document.
// Useful for decoding documents from json/yaml formats.
package document

import (
	"encoding/json"

	"github.com/ogen-go/ogen/jsonschema"
)

// Num represents JSON number.
type Num = jsonschema.Num

// Document is the root document object of the OpenAPI document.
type Document struct {
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

	// Raw JSON value. Used by JSON Schema resolver.
	Raw []byte `json:"-"`
}

func (s *Document) UnmarshalJSON(bytes []byte) error {
	type Alias Document
	var a Alias

	if err := json.Unmarshal(bytes, &a); err != nil {
		return err
	}
	a.Raw = append(a.Raw, bytes...)
	*s = Document(a)

	return nil
}

// Init components of schema.
func (s *Document) Init() {
	if s.Components == nil {
		s.Components = &Components{}
	}

	c := s.Components
	if c.Schemas == nil {
		c.Schemas = make(map[string]*jsonschema.RawSchema)
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
	Schemas         map[string]*jsonschema.RawSchema `json:"schemas,omitempty"`
	Responses       map[string]*Response             `json:"responses,omitempty"`
	Parameters      map[string]*Parameter            `json:"parameters,omitempty"`
	Examples        map[string]*Example              `json:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody          `json:"requestBodies,omitempty"`
	SecuritySchemes map[string]*SecuritySchema       `json:"securitySchemes,omitempty"`

	// Headers         map[string]Header          `json:"headers"`
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
	Tags []string `json:"tags,omitempty"`

	Summary     string               `json:"summary,omitempty"`
	Description string               `json:"description,omitempty"`
	OperationID string               `json:"operationId,omitempty"`
	Parameters  []*Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody         `json:"requestBody,omitempty"`
	Responses   Responses            `json:"responses,omitempty"`
	Security    SecurityRequirements `json:"security,omitempty"`
	Deprecated  bool                 `json:"deprecated,omitempty"`
}

// Parameter describes a single operation parameter.
// A unique parameter is defined by a combination of a name and location.
type Parameter struct {
	Ref  string `json:"$ref,omitempty"`
	Name string `json:"name"`

	// The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In          string                `json:"in"`
	Description string                `json:"description,omitempty"`
	Schema      *jsonschema.RawSchema `json:"schema"`

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
	Schema   *jsonschema.RawSchema `json:"schema,omitempty"`
	Example  json.RawMessage       `json:"example,omitempty"`
	Examples map[string]*Example   `json:"examples,omitempty"`
}

// Discriminator discriminates types for OneOf, AllOf, AnyOf.
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}
