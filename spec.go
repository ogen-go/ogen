package ogen

type Contact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Server struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Path map[string]PathMethod

type PathMethod struct {
	Description string `json:"description"`
	OperationID string `json:"operationId"`
}

type Response struct {
	Description string             `json:"description"`
	Content     map[string]Content `json:"content"`
}

type ContentSchema struct {
	Type  string            `json:"type"`
	Items map[string]string `json:"items"`
}

type Content struct {
	Schema ContentSchema `json:"schema"`
}

type Components struct {
	Schemas map[string]ComponentSchema `json:"schemas"`
}

type ComponentSchema struct {
	Description string                     `json:"description"`
	Type        string                     `json:"type"`
	Format      string                     `json:"format"`
	Properties  map[string]ComponentSchema `json:"properties"`
}

type Spec struct {
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	TermsOfService string          `json:"termsOfService"`
	Contact        *Contact        `json:"contact"`
	License        *License        `json:"license"`
	Version        string          `json:"version"`
	Servers        []Server        `json:"servers"`
	Paths          map[string]Path `json:"paths"`
	Components     *Components     `json:"components"`
}
