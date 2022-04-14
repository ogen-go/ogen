// Package openapi represents OpenAPI v3 Specification in Go.
package openapi

import "github.com/ogen-go/ogen/jsonschema"

type API struct {
	Operations []*Operation
	Components *Components
}

type Components struct {
	Parameters    map[string]*Parameter
	Schemas       map[string]*jsonschema.Schema
	RequestBodies map[string]*RequestBody
	Responses     map[string]*Response
}
