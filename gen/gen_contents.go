package gen

import (
	"fmt"
	"mime"
	"path"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/openapi"
)

func filterMostSpecific(contents map[string]*openapi.MediaType, log *zap.Logger) error {
	keep := func(current, mask string) (string, bool) {
		// Special case for "*", "**", etc.
		var notOnlyStar bool
		for _, c := range mask {
			if c != '*' {
				notOnlyStar = true
				break
			}
		}
		if !notOnlyStar {
			for k := range contents {
				if k == current {
					continue
				}
				// There is at least one another media type, so delete "*".
				return k, false
			}
			// There is no other media type, so keep "*".
			return "", true
		}

		for contentType := range contents {
			// Do not try to match mask against itself.
			if contentType == current {
				continue
			}
			// Found more specific media type that matches the mask, so delete the mask.
			if matched, _ := path.Match(mask, contentType); matched {
				return contentType, false
			}
		}
		// Found no more specific media type, so keep the mask.
		return "", true
	}

	for k := range contents {
		contentType, _, err := mime.ParseMediaType(k)
		if err != nil {
			return errors.Wrapf(err, "parse content type %q", k)
		}

		if replacement, keep := keep(k, contentType); !keep {
			log.Info("Filter common content type",
				zap.String("mask", k),
				zap.String("replacement", replacement),
			)
			delete(contents, k)
		}
	}
	return nil
}

func (g *Generator) wrapContent(ctx *genctx, name string, t *ir.Type) (ret *ir.Type, rerr error) {
	defer func() {
		if rerr != nil {
			return
		}

		if err := ctx.saveType(ret); err != nil {
			rerr = err
			ret = nil
		}
	}()

	if t.Name != "" {
		name = t.Name
	}
	wrapper := &ir.Type{
		Kind: ir.KindStruct,
		Name: name + "WithContentType",
		Doc:  fmt.Sprintf("%sWithContentType wraps %s with Content-Type.", name, t.Go()),
		Fields: []*ir.Field{
			{Name: "ContentType", Type: ir.Primitive(ir.String, nil)},
			{Name: "Content", Type: t},
		},
	}
	return wrapper, nil
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

	var (
		complexTypeErr = func(bt *ir.Type) error {
			impl := &ErrNotImplemented{"complex form schema"}
			if bt != t {
				return errors.Wrapf(impl, "%s -> %s", t, bt)
			}
			return errors.Wrapf(impl, "%s", bt)
		}
		structType = t
	)
	switch t.Kind {
	case ir.KindStruct:
	case ir.KindGeneric:
		generic := t.GenericOf
		if v := t.GenericVariant; optional && v.OnlyOptional() && generic.IsStruct() {
			structType = generic
			break
		}
		return nil, complexTypeErr(generic)
	default:
		return nil, complexTypeErr(t)
	}

	for _, f := range structType.Fields {
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
	optional,
	request bool,
	contents map[string]*openapi.MediaType,
) (_ map[ir.ContentType]ir.Media, err error) {
	if err := filterMostSpecific(contents, g.log); err != nil {
		return nil, errors.Wrap(err, "filter most specific")
	}

	var (
		result = make(map[ir.ContentType]ir.Media, len(contents))
		names  = make(map[ir.ContentType]string, len(contents))

		keys        = xmaps.SortedKeys(contents)
		unsupported []string
		lastErr     error
	)

	for _, contentType := range keys {
		media := contents[contentType]

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
		names[ir.ContentType(parsedContentType)] = typeName

		ctx := ctx.appendPath(contentType)
		if err := func() error {
			encoding := ir.Encoding(parsedContentType)
			if r, ok := g.opt.ContentTypeAliases[parsedContentType]; ok {
				if encoding.MultipartForm() {
					return &ErrNotImplemented{"multipart form CT aliasing"}
				}
				encoding = r
			}

			switch encoding {
			case ir.EncodingJSON:
				t, err := g.generateSchema(ctx.appendPath("schema"), typeName, media.Schema, optional)
				if err != nil {
					return errors.Wrap(err, "generate schema")
				}

				t.AddFeature("json")
				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding: encoding,
					Type:     t,
				}
				return nil

			case ir.EncodingFormURLEncoded:
				t, err := g.generateFormContent(ctx, typeName, media, optional, func(f *ir.Field) error {
					f.Type.AddFeature("uri")
					return nil
				})
				if err != nil {
					return err
				}

				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding: encoding,
					Type:     t,
				}
				return nil

			case ir.EncodingMultipart:
				files := map[string]*ir.Type{}
				t, err := g.generateFormContent(ctx, typeName, media, optional, func(f *ir.Field) error {
					t, err := isMultipartFile(ctx, f.Type, f.Spec)
					if err != nil {
						return err
					}
					if t != nil {
						t.AddFeature("multipart-file")
						files[f.Name] = t
						return nil
					}
					f.Type.AddFeature("uri")
					return nil
				})
				if err != nil {
					return err
				}
				// Create special type for multipart type if form includes file parameters.
				//
				// We need to do it in case when same media definition shared by different content types.
				// For example:
				//
				//	content:
				//    application/json:
				//      schema:
				//        $ref: '#/components/schemas/Form'
				//
				//    multipart/form-data:
				//      schema:
				//        $ref: '#/components/schemas/Form'
				// ...
				//  components:
				//    schemas:
				//      Form:
				//        type: object
				//        properties:
				//          file:
				//            type: string
				//            format: binary
				//
				if len(files) > 0 {
					// TODO(tdakkota): too hacky
					newt := &ir.Type{
						Doc:            t.Doc,
						Kind:           t.Kind,
						Name:           t.Name + "Form",
						Schema:         t.Schema,
						GenericOf:      t.GenericOf,
						GenericVariant: t.GenericVariant,
						Validators:     t.Validators,
					}

					for _, f := range t.Fields {
						fieldType := f.Type
						if file, ok := files[f.Name]; ok {
							fieldType = file
						}
						newt.Fields = append(newt.Fields, &ir.Field{
							Name:   f.Name,
							Type:   fieldType,
							Tag:    f.Tag,
							Inline: f.Inline,
							Spec:   f.Spec,
						})
					}

					if err := ctx.saveType(newt); err != nil {
						return errors.Wrapf(err, "override form %q", t.Name)
					}
					t = newt
				}

				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding: encoding,
					Type:     t,
				}
				return nil

			case ir.EncodingOctetStream:
				if s := media.Schema; s != nil && !isBinary(s) {
					return errors.Wrapf(
						&ErrNotImplemented{Name: "complex application/octet-stream"},
						"generate %q", s.Type,
					)
				}

				// FIXME(tdakkota): box if optional is true?
				t := ir.Stream(typeName)
				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding: encoding,
					Type:     t,
				}
				return ctx.saveType(t)

			case ir.EncodingTextPlain:
				if s := media.Schema; s != nil && s.Type != "string" {
					return errors.Wrapf(
						&ErrNotImplemented{Name: "complex text/plain"},
						"generate %q", s.Type,
					)
				}

				// FIXME(tdakkota): box if optional is true?
				t := ir.Stream(typeName)
				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding: encoding,
					Type:     t,
				}
				return ctx.saveType(t)

			default:
				if isBinary(media.Schema) {
					// FIXME(tdakkota): box if optional is true?
					t := ir.Stream(typeName)
					result[ir.ContentType(parsedContentType)] = ir.Media{
						Encoding: ir.EncodingOctetStream,
						Type:     t,
					}
					return ctx.saveType(t)
				}

				g.log.Info(`Content type is unsupported, set "format" to "binary" to use io.Reader`,
					zapPosition(media),
					zap.String("contentType", contentType),
				)
				unsupported = append(unsupported, contentType)
				return nil
			}
		}(); err != nil {
			err = errors.Wrapf(err, "media: %q", contentType)
			if err := g.trySkip(err, "Skipping media", media); err != nil {
				return nil, err
			}
			lastErr = err
			unsupported = append(unsupported, contentType)
			continue
		}
	}

	if len(result) == 0 && len(unsupported) > 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, &ErrUnsupportedContentTypes{ContentTypes: unsupported}
	}

	if request {
		for ct, m := range result {
			if !ct.Mask() {
				continue
			}
			t, err := g.wrapContent(ctx, names[ct], m.Type)
			if err != nil {
				return nil, err
			}
			result[ct] = ir.Media{
				Type:     t,
				Encoding: m.Encoding,
			}
		}
	}

	return result, nil
}
