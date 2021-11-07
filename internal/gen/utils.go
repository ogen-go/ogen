package gen

import (
	"sort"

	"github.com/ogen-go/ogen/internal/oas"
)

func sortContentTypes(contents map[string]*oas.Schema) ([]string, error) {
	contentTypes := make([]string, 0, len(contents))
	unsupported := make([]string, 0)
	for contentType, schema := range contents {
		switch contentType {
		case "application/json", "application/octet-stream":
			contentTypes = append(contentTypes, contentType)
		default:
			if isBinary(schema) {
				contentTypes = append(contentTypes, contentType)
				continue
			}

			unsupported = append(unsupported, contentType)
			continue
		}
	}

	if len(contentTypes) == 0 && len(unsupported) > 0 {
		return nil, &ErrUnsupportedContentTypes{ContentTypes: unsupported}
	}

	sort.Strings(contentTypes)
	return contentTypes, nil
}

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
