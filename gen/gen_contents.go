package gen

import (
	"mime"
	"path"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func filterMostSpecific(contents map[string]*openapi.MediaType) error {
	initialLength := len(contents)
	keep := func(current, mask string) bool {
		// Special case for "*", "**", etc.
		var nonStar bool
		for _, c := range mask {
			if c != '*' {
				nonStar = true
				break
			}
		}
		if !nonStar {
			return initialLength < 2
		}

		for contentType := range contents {
			if contentType == current {
				continue
			}
			if matched, _ := path.Match(mask, contentType); matched {
				return false
			}
		}
		return true
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

func (g *Generator) generateContents(
	ctx *genctx,
	name string,
	optional bool,
	contents map[string]*openapi.MediaType,
) (_ map[ir.ContentType]*ir.Type, err error) {
	var (
		result      = make(map[ir.ContentType]*ir.Type, len(contents))
		unsupported []string
	)
	if err := filterMostSpecific(contents); err != nil {
		return nil, errors.Wrap(err, "filter most specific")
	}

	for contentType, media := range contents {
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
				t, err := g.generateSchema(ctx, typeName, media.Schema)
				if err != nil {
					return errors.Wrap(err, "generate schema")
				}

				t.AddFeature("json")
				t, err = boxType(ctx, ir.GenericVariant{
					Nullable: t.Schema != nil && t.Schema.Nullable,
					Optional: optional,
				}, t)
				if err != nil {
					return errors.Wrap(err, "box schema")
				}

				result[ir.ContentTypeJSON] = t
				return nil

			case "application/x-www-form-urlencoded":
				t, err := g.generateSchema(ctx, typeName, media.Schema)
				if err != nil {
					return errors.Wrap(err, "generate schema")
				}
				if !t.IsStruct() {
					return errors.Wrapf(&ErrNotImplemented{"urlencoded schema type"}, "%s", t.Kind)
				}

				t.AddFeature("urlencoded")
				t, err = boxType(ctx, ir.GenericVariant{
					Nullable: t.Schema != nil && t.Schema.Nullable,
					Optional: optional,
				}, t)
				if err != nil {
					return errors.Wrap(err, "box schema")
				}

				result[ir.ContentTypeFormURLEncoded] = t
				return nil

			case "application/octet-stream":
				if media.Schema != nil && !isBinary(media.Schema) {
					return errors.Errorf("octet stream with %q schema not supported", media.Schema.Type)
				}

				t := ir.Stream(typeName)
				result[ir.ContentTypeOctetStream] = t
				return ctx.saveType(t)

			default:
				if isBinary(media.Schema) {
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
