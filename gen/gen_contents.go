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
	optional bool,
	contents map[string]*openapi.MediaType,
) (_ map[ir.ContentType]ir.Media, err error) {
	var (
		result      = make(map[ir.ContentType]ir.Media, len(contents))
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
			encoding := ir.Encoding(parsedContentType)
			if r, ok := g.opt.ContentTypeAliases[parsedContentType]; ok {
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

				t := ir.Stream(typeName)
				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding: encoding,
					Type:     t,
				}
				return ctx.saveType(t)

			default:
				if isBinary(media.Schema) {
					t := ir.Stream(typeName)
					result[ir.ContentType(parsedContentType)] = ir.Media{
						Encoding: ir.EncodingOctetStream,
						Type:     t,
					}
					return ctx.saveType(t)
				}

				g.log.Info(`Content type is unsupported, set "format" to "binary" to use io.Reader`,
					g.zapLocation(media),
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
