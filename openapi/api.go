// Package openapi represents OpenAPI v3 Specification in Go.
package openapi

import "github.com/ogen-go/ogen/jsonschema"

// API represents parsed OpenAPI spec.
type API struct {
	Servers    []Server
	Operations []*Operation
	Components *Components
}

// Server represents parsed OpenAPI Server Object.
type Server struct {
	Template    Path
	Description string // optional
}

// Components represent parsed components of OpenAPI spec.
type Components struct {
	Schemas       map[string]*jsonschema.Schema
	Responses     map[string]*Response
	Parameters    map[string]*Parameter
	Examples      map[string]*Example
	RequestBodies map[string]*RequestBody
}
