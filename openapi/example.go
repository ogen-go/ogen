package openapi

import (
	"encoding/json"

	ogenjson "github.com/ogen-go/ogen/json"
)

// Example is an OpenAPI Example.
type Example struct {
	Ref string

	Summary       string
	Description   string
	Value         json.RawMessage
	ExternalValue string

	ogenjson.Locator
}
