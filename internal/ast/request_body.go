package ast

type RequestBody struct {
	Contents map[string]*Schema
	Required bool
}
