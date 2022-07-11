package openapi

import (
	ogenjson "github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/jsonschema"
)

// Example is an OpenAPI Example.
type Example struct {
	Ref string

	Summary       string
	Description   string
	Value         jsonschema.Example
	ExternalValue string

	ogenjson.Locator `json:"-" yaml:"-"`
}
