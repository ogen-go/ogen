package ir

import (
	"strings"

	"github.com/ogen-go/ogen/openapi"
)

// Servers is a list of servers.
type Servers []Server

func (s Servers) filter(cb func(Server) bool) (r []Server) {
	for _, server := range s {
		if cb(server) {
			r = append(r, server)
		}
	}
	return r
}

// Templates returns a list of server URL templates.
func (s Servers) Templates() []Server {
	return s.filter(Server.IsTemplate)
}

// Const return a list of constant server URLs.
func (s Servers) Const() []Server {
	return s.filter(func(server Server) bool {
		return !server.IsTemplate()
	})
}

// Server describes a OpenAPI server.
type Server struct {
	Name   string
	Params []ServerParam
	Spec   openapi.Server
}

// IsTemplate returns true if server URL has variables.
func (s Server) IsTemplate() bool {
	return len(s.Params) > 0
}

// ServerParam describes a server template parameter.
type ServerParam struct {
	// Name is a Go name of the parameter.
	Name string
	Spec openapi.ServerVariable
}

// FormatString returns a format string (fmt.Sprintf) for the server.
//
// If the server has no variables, returns plain string.
func (s Server) FormatString() string {
	var sb strings.Builder
	for _, part := range s.Spec.Template {
		if part.IsParam() {
			sb.WriteString("%s")
		} else {
			sb.WriteString(part.Raw)
		}
	}
	return sb.String()
}

// GoDoc returns GoDoc comment for the server.
func (s Server) GoDoc() []string {
	return prettyDoc(s.Spec.Description, "")
}
