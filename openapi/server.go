package openapi

// Server represents parsed OpenAPI Server Object.
type Server struct {
	Name        string // optional,extension
	Description string // optional
	Template    ServerURL
}

// ServerVariable represents parsed OpenAPI Server Variable Object.
type ServerVariable struct {
	Name        string
	Description string
	Default     string
	Enum        []string
}

// ServerURL is URL template with variables.
type ServerURL []PathPart[ServerVariable]
