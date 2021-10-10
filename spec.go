package ogen

import "encoding/json"

// This is the root document object of the OpenAPI document.
type Spec struct {
	// This string MUST be the semantic version number
	// of the OpenAPI Specification version that the OpenAPI document uses.
	OpenAPI    string      `json:"openapi"`
	Info       Info        `json:"info"`
	Servers    []Server    `json:"servers"`
	Paths      Paths       `json:"paths"`
	Components *Components `json:"components"`
}

// The object provides metadata about the API.
// The metadata MAY be used by the clients if needed,
// and MAY be presented in editing or documentation generation tools for convenience.
type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title"`
	// A short description of the API.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`
	// A URL to the Terms of Service for the API. MUST be in the format of a URL.
	TermsOfService string `json:"termsOfService"`
	// The contact information for the exposed API.
	Contact *Contact `json:"contact"`
	// The license information for the exposed API.
	License *License `json:"license"`
	// REQUIRED. The version of the OpenAPI document.
	Version string `json:"version"`
}

// Contact information for the exposed API.
type Contact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

// License information for the exposed API.
type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// An object representing a Server.
type Server struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Holds a set of reusable objects for different aspects of the OAS.
// All objects defined within the components object will have no effect on the API
// unless they are explicitly referenced from properties outside the components object.
type Components struct {
	Schemas    map[string]Schema    `json:"schemas"`
	Responses  map[string]Response  `json:"responses"`
	Parameters map[string]Parameter `json:"parameters"`
	// Examples        map[string]Example         `json:"example"`
	RequestBodies map[string]RequestBody `json:"requestBodies"`
	// Headers         map[string]Header          `json:"headers"`
	// SecuritySchemes map[string]SecuritySchema  `json:"securitySchemes"`
	// Links           map[string]Link            `json:"links"`
	// Callbacks       map[string]Callback        `json:"callback"`
}

// Paths holds the relative paths to the individual endpoints and their operations.
// The path is appended to the URL from the Server Object in order to construct the full URL.
// The Paths MAY be empty, due to ACL constraints.
type Paths map[string]PathItem

// PathItem describes the operations available on a single path.
// A Path Item MAY be empty, due to ACL constraints.
// The path itself is still exposed to the documentation viewer
// but they will not know which operations and parameters are available.
type PathItem struct {
	// Allows for an external definition of this path item.
	// The referenced structure MUST be in the format of a Path Item Object.
	// In case a Path Item Object field appears both
	// in the defined object and the referenced object, the behavior is undefined.
	Ref         string      `json:"$ref"`
	Description string      `json:"description,omitempty"`
	Get         *Operation  `json:"get"`
	Put         *Operation  `json:"put"`
	Post        *Operation  `json:"post"`
	Delete      *Operation  `json:"delete"`
	Options     *Operation  `json:"options"`
	Head        *Operation  `json:"head"`
	Patch       *Operation  `json:"patch"`
	Trace       *Operation  `json:"trace"`
	Servers     []Server    `json:"servers"`
	Parameters  []Parameter `json:"parameters"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	// A list of tags for API documentation control.
	// Tags can be used for logical grouping of operations by resources or any other qualifier.
	Tags        []string     `json:"tags,omitempty"`
	Description string       `json:"description,omitempty"`
	OperationID string       `json:"operationId"`
	Parameters  []Parameter  `json:"parameters"`
	RequestBody *RequestBody `json:"requestBody"`
	Responses   Responses    `json:"responses"`
}

// Describes a single operation parameter.
// A unique parameter is defined by a combination of a name and location.
type Parameter struct {
	Ref  string `json:"$ref"`
	Name string `json:"name"`

	// The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In          string `json:"in"`
	Description string `json:"description"`
	Schema      Schema `json:"schema"`

	// Determines whether this parameter is mandatory.
	// If the parameter location is "path", this property is REQUIRED
	// and its value MUST be true.
	// Otherwise, the property MAY be included and its default value is false.
	Required bool `json:"required"`

	// Specifies that a parameter is deprecated and SHOULD be transitioned out of usage.
	// Default value is false.
	Deprecated bool `json:"deprecated"` // TODO: implement

	// For more complex scenarios, the content property can define the media type and schema of the parameter.
	// A parameter MUST contain either a schema property, or a content property, but not both.
	// When example or examples are provided in conjunction with the schema object,
	// the example MUST follow the prescribed serialization strategy for the parameter.
	//
	// A map containing the representations for the parameter.
	// The key is the media type and the value describes it.
	// The map MUST only contain one entry.
	Content map[string]Media `json:"content"` // TODO: implement

	// Describes how the parameter value will be serialized
	// depending on the type of the parameter value.
	Style string `json:"style"`

	// When this is true, parameter values of type array or object
	// generate separate parameters for each value of the array
	// or key-value pair of the map.
	// For other types of parameters this property has no effect.
	Explode *bool `json:"explode"`
}

// RequestBody describes a single request body.
type RequestBody struct {
	Ref         string `json:"$ref"`
	Description string `json:"description"`

	// The content of the request body.
	// The key is a media type or media type range and the value describes it.
	// For requests that match multiple keys, only the most specific key is applicable.
	// e.g. text/plain overrides text/*
	Content map[string]Media `json:"content"`

	// Determines if the request body is required in the request. Defaults to false.
	Required bool `json:"required"`
}

// Responses - a container for the expected responses of an operation.
// The container maps a HTTP response code to the expected response
type Responses map[string]Response

// Describes a single response from an API Operation,
// including design-time, static links to operations based on the response.
type Response struct {
	Ref         string                 `json:"$ref"`
	Description string                 `json:"description"`
	Header      map[string]interface{} // TODO: implement
	Content     map[string]Media       `json:"content"`
	Links       map[string]interface{} // TODO: implement
}

// Media provides schema and examples for the media type identified by its key.
type Media struct {
	// The schema defining the content of the request, response, or parameter.
	Schema Schema `json:"schema"`
}

// The Schema Object allows the definition of input and output data types.
// These types can be objects, but also primitives and arrays.
type Schema struct {
	Ref         string `json:"$ref"`
	Description string `json:"description"`

	// Value MUST be a string. Multiple types via an array are not supported.
	Type string `json:"type"`

	// See Data Type Formats for further details (https://swagger.io/specification/#data-type-format).
	// While relying on JSON Schema's defined formats,
	// the OAS offers a few additional predefined formats.
	Format string `json:"format"`

	// Property definitions MUST be a Schema Object and not a standard JSON Schema
	// (inline or referenced).
	Properties map[string]Schema `json:"properties"`

	// The value of this keyword MUST be an array.
	// This array MUST have at least one element.
	// Elements of this array MUST be strings, and MUST be unique.
	Required []string `json:"required"`

	// Value MUST be an object and not an array.
	// Inline or referenced schema MUST be of a Schema Object and not a standard
	Items *Schema `json:"items"`

	// AllOf takes an array of object definitions that are used
	// for independent validation but together compose a single object.
	// Still, it does not imply a hierarchy between the models.
	// For that purpose, you should include the discriminator.
	AllOf []Schema `json:"allOf"` // TODO: implement.

	// OneOf validates the value against exactly one of the subschemas
	OneOf []Schema `json:"oneOf"` // TODO: implement.

	// AnyOf validates the value against any (one or more) of the subschemas
	AnyOf []Schema `json:"anyOf"` // TODO: implement.

	// The value of this keyword MUST be an array.
	// This array SHOULD have at least one element.
	// Elements in the array SHOULD be unique.
	Enum []json.RawMessage `json:"enum"` // TODO: Nullable.
}
