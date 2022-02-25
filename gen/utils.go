package gen

import (
	"fmt"
	"net/http"

	"github.com/ogen-go/ogen/jsonschema"
)

func isBinary(s *jsonschema.Schema) bool {
	if s == nil {
		return false
	}

	switch s.Type {
	case jsonschema.Empty, jsonschema.String:
		return s.Format == "binary"
	default:
		return false
	}
}

func statusText(code int) string {
	r := http.StatusText(code)
	if r != "" {
		return r
	}
	return fmt.Sprintf("Code%d", code)
}
