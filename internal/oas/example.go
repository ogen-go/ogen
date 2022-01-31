package oas

import "encoding/json"

// Example is an OpenAPI Example.
type Example struct {
	Ref           string          `json:"$ref,omitempty"` // ref object
	Summary       string          `json:"summary,omitempty"`
	Description   string          `json:"description,omitempty"`
	Value         json.RawMessage `json:"value,omitempty"`
	ExternalValue string          `json:"externalValue,omitempty"`
}
