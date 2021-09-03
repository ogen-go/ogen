package ogen

type Spec struct {
	// This string MUST be the semantic version number
	// of the OpenAPI Specification version that the OpenAPI document uses.
	OpenAPI        string      `json:"openapi"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	TermsOfService string      `json:"termsOfService"`
	Contact        *Contact    `json:"contact"`
	License        *License    `json:"license"`
	Version        string      `json:"version"`
	Servers        []Server    `json:"servers"`
	Paths          Paths       `json:"paths"`
	Components     *Components `json:"components"`
}

type Contact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Server struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas"`
	// Responses       map[string]Response        `json:"responses"`
	// Parameters      map[string]Parameter       `json:"parameters"`
	// Examples        map[string]Example         `json:"example"`
	// RequiesBodies   map[string]RequestBody     `json:"requestBodies"`
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
	Tags        []string    `json:"tags,omitempty"`
	Description string      `json:"description,omitempty"`
	OperationID string      `json:"operationId"`
	Parameters  []Parameter `json:"parameters"`
	RequestBody RequestBody `json:"requestBody"`
	Responses   Responses   `json:"responses"`
}

type Parameter struct {
	Name string `json:"name"`
	// The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In          string `json:"in"`
	Description string `json:"description"`
	Schema      Schema `json:"schema"`
	// Determines whether this parameter is mandatory.
	// If the parameter location is "path", this property is REQUIRED
	// and its value MUST be true.
	// Otherwise, the property MAY be included and its default value is false.
	Required   bool `json:"required"`
	Deprecated bool `json:"deprecated"` // TODO: implement
}

type RequestBody struct {
	Description string             `json:"description"`
	Content     map[string]Content `json:"content"`
	Required    bool               `json:"required"` // TODO: implement
}

// Responses - a container for the expected responses of an operation.
// The container maps a HTTP response code to the expected response
type Responses map[string]Response

type Response struct {
	Description string                 `json:"description"`
	Header      map[string]interface{} // TODO: implement
	Content     map[string]Content     `json:"content"`
	Links       map[string]interface{} // TODO: implement
}

type ContentSchema struct {
	Type  string            `json:"type"`
	Items map[string]string `json:"items"`
	Ref   string            `json:"$ref"`
}

type Content struct {
	Schema ContentSchema `json:"schema"`
}

type Schema struct {
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
	// Value MUST be an object and not an array. Inline or referenced schema MUST be of a Schema Object and not a standard
	Items *Schema `json:"items"`
	Ref   string  `json:"$ref"`
}
