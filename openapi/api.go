// Package openapi represents OpenAPI v3 Specification in Go.
package openapi

import "github.com/ogen-go/ogen/jsonschema"

// API represents parsed OpenAPI spec.
type API struct {
	Operations []*Operation
	Components *Components
}

// Components represent parsed components of OpenAPI spec.
type Components struct {
	Parameters    map[string]*Parameter
	Schemas       map[string]*jsonschema.Schema
	RequestBodies map[string]*RequestBody
	Responses     map[string]*Response
}
