package gen

import "github.com/ogen-go/ogen/jsonschema"

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
