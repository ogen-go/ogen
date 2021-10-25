package ast

import "strings"

type ParameterLocation string

const (
	LocationQuery  ParameterLocation = "Query"
	LocationHeader ParameterLocation = "Header"
	LocationPath   ParameterLocation = "Path"
	LocationCookie ParameterLocation = "Cookie"
)

func (p ParameterLocation) Lower() string { return strings.ToLower(string(p)) }

type Parameter struct {
	Name     string
	Schema   *Schema
	In       ParameterLocation
	Style    string
	Explode  bool
	Required bool
}
