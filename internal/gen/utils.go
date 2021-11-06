package gen

import (
	"sort"

	"github.com/ogen-go/ogen/internal/oas"
)

func sortContentTypes(contents map[string]*oas.Schema) ([]string, error) {
	contentTypes := make([]string, 0, len(contents))
	unsupported := make([]string, 0)
	for contentType := range contents {
		switch contentType {
		case "application/json", "application/octet-stream":
			contentTypes = append(contentTypes, contentType)
		default:
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
