package gen

import (
	"sort"

	"github.com/ogen-go/ogen/internal/oas"
)

func sortContentTypes(contents map[string]*oas.Schema) ([]string, error) {
	contentTypes := make([]string, 0, len(contents))
	unsupported := make([]string, 0)
	for contentType := range contents {
		if contentType != "application/json" {
			unsupported = append(unsupported, contentType)
			continue
		}
		contentTypes = append(contentTypes, contentType)
	}

	if len(contentTypes) == 0 && len(unsupported) > 0 {
		return nil, &ErrUnsupportedContentTypes{ContentTypes: unsupported}
	}

	sort.Strings(contentTypes)
	return contentTypes, nil
}
