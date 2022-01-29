package gen

import (
	"github.com/ogen-go/ogen/internal/oas"
)

func isBinary(s *oas.Schema) bool {
	if s == nil {
		return false
	}

	switch s.Type {
	case "", oas.String:
	default:
		return false
	}

	return s.Format == oas.FormatBinary
}
