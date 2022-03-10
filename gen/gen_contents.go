package gen

import (
	"mime"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func filterMostSpecific(contents map[string]*jsonschema.Schema) error {
	keep := func(current, mask string) bool {
		for {
			star := strings.IndexByte(mask, '*')
			if star < 0 {
				return true
			}

			prefix := mask[:star]
			for contentType := range contents {
				if contentType == current {
					continue
				}
				if strings.HasPrefix(contentType, prefix) {
					return false
				}
			}
			mask = mask[star+1:]
		}
	}

	for k := range contents {
		contentType, _, err := mime.ParseMediaType(k)
		if err != nil {
			return errors.Wrapf(err, "parse content type %q", k)
		}

		if !keep(k, contentType) {
			delete(contents, k)
		}
	}
	return nil
}

func (g *Generator) generateContents(ctx *genctx, name string, optional bool, contents map[string]*jsonschema.Schema) (_ map[ir.ContentType]*ir.Type, err error) {
	var (
		result      = make(map[ir.ContentType]*ir.Type, len(contents))
		unsupported []string
	)
	if err := filterMostSpecific(contents); err != nil {
		return nil, errors.Wrap(err, "filter most specific")
	}

	for contentType, schema := range contents {
		parsedContentType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			return nil, errors.Wrapf(err, "parse content type %q", contentType)
		}

		typeName := name
		if len(contents) > 1 {
			typeName, err = pascal(name, contentType)
			if err != nil {
				return nil, errors.Wrapf(err, "name for %q", contentType)
			}
		}

		ctx := ctx.appendPath(contentType)
		if err := func() error {
			switch parsedContentType {
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
					return errors.Errorf("octet stream with %q schema not supported", schema.Type)
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
