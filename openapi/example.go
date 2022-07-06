package openapi

import (
	"github.com/go-json-experiment/json"

	ogenjson "github.com/ogen-go/ogen/json"
)

// Example is an OpenAPI Example.
type Example struct {
	Ref string

	Summary       string
	Description   string
	Value         json.RawValue
	ExternalValue string

	ogenjson.Locator
}
