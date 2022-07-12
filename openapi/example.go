package openapi

import (
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/jsonschema"
)

// Example is an OpenAPI Example.
type Example struct {
	Ref string

	Summary       string
	Description   string
	Value         jsonschema.Example
	ExternalValue string

	location.Locator `json:"-" yaml:"-"`
}
