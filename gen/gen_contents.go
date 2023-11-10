package gen

import (
	"fmt"
	"mime"
	"path"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/jsonschema"
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
	encoding ir.Encoding,
) (*ir.Type, error) {
	if s := media.Schema; s != nil &&
		((s.AdditionalProperties != nil && s.Item != nil) ||
			len(s.PatternProperties) > 0) ||
		len(s.Items) > 0 {
		return nil, &ErrNotImplemented{"complex form schema"}
	}

	getEncoding := func(f *ir.Field) (ct ir.Encoding) {
		if e, ok := media.Encoding[f.Tag.JSON]; ok {
			ct = ir.Encoding(e.ContentType)
		}
		if ct == "" && encoding.MultipartForm() && isComplexMultipartType(f.Spec.Schema) {
			ct = ir.EncodingJSON
		}
		return ct
	}

	var override generateSchemaOverride
	switch encoding {
	case ir.EncodingFormURLEncoded:
		override.fieldMut = func(f *ir.Field) error {
			f.Type.AddFeature("uri")
			return nil
		}
	case ir.EncodingMultipart:
		// A funny moment when you have a spec that shares schema between multipart form and JSON request and
		// at some point you made ingenious decision to keep all types in one package at the same time.
		if s := media.Schema; s != nil && !s.Ref.IsZero() {
			override.refEncoding = map[jsonschema.Ref]ir.Encoding{
				s.Ref: encoding,
			}
			override.nameRef = func(ref jsonschema.Ref, def refNamer) (string, error) {
				n, err := def(ref)
				if err == nil && ref == s.Ref {
					n += "Multipart"
				}
				return n, err
			}
		}
		override.fieldMut = func(f *ir.Field) error {
			t, err := isMultipartFile(ctx, f.Type, f.Spec)
			if err != nil {
				return err
			}
			if t != nil {
				f.Type = t
				t.AddFeature("multipart-file")
				return nil
			}
			switch ct := getEncoding(f); ct {
			case "", ir.EncodingFormURLEncoded:
				f.Type.AddFeature("uri")
			case ir.EncodingJSON:
				f.Type.AddFeature("json")
			default:
				return errors.Wrapf(
					&ErrNotImplemented{"form content encoding"},
					"%q", ct,
				)
			}
			return nil
		}
	}
	t, err := g.generateSchema(ctx, typeName, media.Schema, optional, &override)
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
			}
			switch ct := getEncoding(f); ct {
			case "", ir.EncodingFormURLEncoded:
				if err := isSupportedParamStyle(spec); err != nil {
					return err
				}

				if err := isParamAllowed(f.Type, true, map[*ir.Type]struct{}{}); err != nil {
					return err
				}
			case ir.EncodingJSON:
				spec.Content = &openapi.ParameterContent{
					Name: ct.String(),
				}
			default:
				return errors.Wrapf(
					&ErrNotImplemented{"form content encoding"},
					"%q", ct,
				)
			}

			return nil
		}(); err != nil {
			return nil, errors.Wrapf(err, "form parameter %q", tag)
		}

		f.Tag.Form = spec
	}
	return t, nil
}

func isComplexMultipartType(s *jsonschema.Schema) bool {
	if s == nil {
		return true
	}

	switch s.Type {
	case jsonschema.Object, jsonschema.Empty:
		return true
	case jsonschema.Array:
		return len(s.Items) > 0 || isComplexMultipartType(s.Item)
	default:
		return false
	}
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

		if err := func() error {
			encoding := ir.Encoding(parsedContentType)
			if r, ok := g.opt.ContentTypeAliases[parsedContentType]; ok {
				if encoding.MultipartForm() {
					return &ErrNotImplemented{"multipart form CT aliasing"}
				}
				encoding = r
			}

			if encoding != ir.EncodingJSON && media.XOgenJSONStreaming {
				g.log.Warn(`Extension "x-ogen-json-streaming" will be ignored for non-JSON encoding`,
					zapPosition(media),
					zap.String("contentType", contentType),
				)
			}

			switch encoding {
			case ir.EncodingJSON:
				t, err := g.generateSchema(ctx, typeName, media.Schema, optional, nil)
				if err != nil {
					return errors.Wrap(err, "generate schema")
				}

				t.AddFeature("json")
				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding:      encoding,
					Type:          t,
					JSONStreaming: media.XOgenJSONStreaming,
				}
				return nil

			case ir.EncodingFormURLEncoded:
				t, err := g.generateFormContent(ctx, typeName, media, optional, encoding)
				if err != nil {
					return err
				}

				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding: encoding,
					Type:     t,
				}
				return nil

			case ir.EncodingMultipart:
				t, err := g.generateFormContent(ctx, typeName, media, optional, encoding)
				if err != nil {
					return err
				}

				result[ir.ContentType(parsedContentType)] = ir.Media{
					Encoding: encoding,
					Type:     t,
				}
				return nil
			default:
				if s := media.Schema; isStream(s) {
					// FIXME(tdakkota): box if optional is true?
					t := ir.Stream(typeName, s)

					switch encoding {
					case ir.EncodingOctetStream, ir.EncodingTextPlain:
					default:
						encoding = ir.EncodingOctetStream
					}
					result[ir.ContentType(parsedContentType)] = ir.Media{
						Encoding: encoding,
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
				Type:          t,
				Encoding:      m.Encoding,
				JSONStreaming: m.JSONStreaming,
			}
		}
	}

	return result, nil
}
