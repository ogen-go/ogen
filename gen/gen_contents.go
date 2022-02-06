package gen

import (
	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/jsonschema"
)

func (g *Generator) generateContents(name string, contents map[string]*jsonschema.Schema) (map[ir.ContentType]*ir.Type, error) {
	var (
		result      = make(map[ir.ContentType]*ir.Type, len(contents))
		unsupported []string
	)

	for contentType, schema := range contents {
		typeName := name
		if len(contents) > 1 {
			typeName = pascal(name, contentType)
		}

		switch contentType {
		case "application/json":
			t, err := g.generateSchema(typeName, schema)
			if err != nil {
				return nil, errors.Wrap(err, "schema")
			}

			t.AddFeature("json")
			result[ir.ContentTypeJSON] = t

		case "application/octet-stream":
			if schema != nil {
				return nil, errors.Errorf("octet stream with schema not supported")
			}

			t := ir.Stream(typeName)
			result[ir.ContentTypeOctetStream] = t
			g.saveType(t)
			continue

		default:
			if isBinary(schema) {
				t := ir.Stream(typeName)
				result[ir.ContentType(contentType)] = t
				g.saveType(t)
				continue
			}

			unsupported = append(unsupported, contentType)
		}
	}

	if len(result) == 0 && len(unsupported) > 0 {
		return nil, &ErrUnsupportedContentTypes{ContentTypes: unsupported}
	}

	return result, nil
}
