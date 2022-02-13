package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func (g *Generator) generateContents(ctx *genctx, name string, optional bool, contents map[string]*jsonschema.Schema) (map[ir.ContentType]*ir.Type, error) {
	var (
		result      = make(map[ir.ContentType]*ir.Type, len(contents))
		unsupported []string
	)

	for contentType, schema := range contents {
		typeName := name
		if len(contents) > 1 {
			typeName = pascal(name, contentType)
		}

		ctx := ctx.appendPath(contentType)
		if err := func() error {
			switch contentType {
			case "application/json":
				t, err := g.generateSchema(ctx, typeName, schema)
				if err != nil {
					return errors.Wrap(err, "schema")
				}

				t.AddFeature("json")
				t, err = boxType(ctx, ir.GenericVariant{
					Nullable: t.Schema != nil && t.Schema.Nullable,
					Optional: optional,
				}, t)
				if err != nil {
					return errors.Wrap(err, "schema")
				}

				result[ir.ContentTypeJSON] = t
				return nil

			case "application/octet-stream":
				if schema != nil && !isBinary(schema) {
					return errors.Errorf("octet stream with schema not supported")
				}

				t := ir.Stream(typeName)
				result[ir.ContentTypeOctetStream] = t
				return ctx.saveType(t)

			default:
				if isBinary(schema) {
					t := ir.Stream(typeName)
					result[ir.ContentType(contentType)] = t
					return ctx.saveType(t)
				}

				unsupported = append(unsupported, contentType)
				return nil
			}
		}(); err != nil {
			return nil, errors.Wrap(err, contentType)
		}
	}

	if len(result) == 0 && len(unsupported) > 0 {
		return nil, &ErrUnsupportedContentTypes{ContentTypes: unsupported}
	}

	return result, nil
}
