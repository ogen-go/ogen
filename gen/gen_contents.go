package gen

import (
	"mime"
	"path"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

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

func (g *Generator) generateFormContent(
	ctx *genctx,
	typeName string,
	media *openapi.MediaType,
	optional bool,
	cb func(f *ir.Field) error,
) (*ir.Type, error) {
	if s := media.Schema; s != nil && (s.AdditionalProperties != nil || len(s.PatternProperties) > 0) {
		return nil, &ErrNotImplemented{"complex form schema"}
	}

	t, err := g.generateSchema(ctx.appendPath("schema"), typeName, media.Schema, optional)
	if err != nil {
		return nil, errors.Wrap(err, "generate schema")
	}
	if !t.IsStruct() {
		return nil, errors.Wrapf(&ErrNotImplemented{"complex form schema"}, "%s", t.Kind)
	}

	for _, f := range t.Fields {
		tag := f.Tag.JSON

		spec := &openapi.Parameter{
			Name:     tag,
			Schema:   f.Spec.Schema,
			In:       openapi.LocationQuery,
			Style:    openapi.QueryStyleForm,
			Explode:  true,
			Required: f.Spec.Required,
		}

		if err := func() error {
			if e, ok := media.Encoding[tag]; ok {
				spec.Style = e.Style
				spec.Explode = e.Explode
				if e.ContentType != "" {
					return &ErrNotImplemented{"parameter content-type"}
				}
			}

			if err := cb(f); err != nil {
				return err
			}

			if err := isSupportedParamStyle(spec); err != nil {
				return err
			}

			if err := isParamAllowed(f.Type, true, map[*ir.Type]struct{}{}); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			return nil, errors.Wrapf(err, "form parameter %q", tag)
		}

		f.Tag.Form = spec
	}
	return t, nil
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
				t, err := g.generateSchema(ctx.appendPath("schema"), typeName, media.Schema, optional)
				if err != nil {
					return errors.Wrap(err, "generate schema")
				}

				t.AddFeature("json")
				result[ir.ContentTypeJSON] = t
				return nil

			case "application/x-www-form-urlencoded":
				t, err := g.generateFormContent(ctx, typeName, media, optional, func(f *ir.Field) error {
					f.Type.AddFeature("uri")
					return nil
				})
				if err != nil {
					return err
				}

				result[ir.ContentTypeFormURLEncoded] = t
				return nil

			case "multipart/form-data":
				t, err := g.generateFormContent(ctx, typeName, media, optional, func(f *ir.Field) error {
					if s := f.Spec; s != nil && isBinary(s.Schema) {
						if !s.Required {
							return &ErrNotImplemented{"optional multipart file"}
						}
						f.Type = ir.Primitive(ir.File, nil)
						return nil
					}
					f.Type.AddFeature("uri")
					return nil
				})
				if err != nil {
					return err
				}

				result[ir.ContentTypeMultipart] = t
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

				g.log.Info(`Content type is unsupported, set "format" to "binary" to use io.Reader`,
					zap.String("contentType", contentType),
				)
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
