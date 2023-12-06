package parser

import (
	"encoding/json"
	"mime"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseContent(
	content map[string]ogen.Media,
	locator location.Locator,
	ctx *jsonpointer.ResolveCtx,
) (_ map[string]*openapi.MediaType, err error) {
	if len(content) == 0 {
		return nil, nil
	}

	result := make(map[string]*openapi.MediaType, len(content))
	for name, m := range content {
		ct, _, err := mime.ParseMediaType(name)
		if err != nil {
			err := errors.Wrapf(err, "content type %q", name)
			return nil, p.wrapLocation(p.file(ctx), locator.Key(name), err)
		}

		result[name], err = p.parseMediaType(ct, m, ctx)
		if err != nil {
			return nil, errors.Wrap(err, name)
		}
	}

	return result, nil
}

func (p *parser) parseParameterContent(
	content map[string]ogen.Media,
	locator location.Locator,
	ctx *jsonpointer.ResolveCtx,
) (*openapi.ParameterContent, error) {
	if content == nil {
		return nil, nil
	}
	if len(content) != 1 {
		me := new(location.MultiError)
		n := 0
		for key := range content {
			// Report only 2 entries.
			if n >= 2 {
				break
			}

			// Set message for the first position.
			var msg string
			if n == 0 {
				msg = `"content" map MUST only contain one entry`
			}

			me.Report(p.file(ctx), locator.Key(key), msg)
			n++
		}
		return nil, me
	}

	for name, m := range content {
		ct, _, err := mime.ParseMediaType(name)
		if err != nil {
			err := errors.Wrapf(err, "content type %q", name)
			return nil, p.wrapLocation(p.file(ctx), locator.Key(name), err)
		}

		media, err := p.parseMediaType(ct, m, ctx)
		if err != nil {
			return nil, errors.Wrap(err, name)
		}

		return &openapi.ParameterContent{
			Name:  name,
			Media: media,
		}, nil
	}
	panic("unreachable")
}

func (p *parser) parseMediaType(ct string, m ogen.Media, ctx *jsonpointer.ResolveCtx) (_ *openapi.MediaType, rerr error) {
	locator := m.Common.Locator
	defer func() {
		rerr = p.wrapLocation(p.file(ctx), locator, rerr)
	}()

	s, err := p.parseSchema(m.Schema, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schema")
	}

	encodings := make(map[string]*openapi.Encoding, len(m.Encoding))
	if len(m.Encoding) > 0 {
		switch ct {
		case "application/x-www-form-urlencoded", "multipart/form-data":
			var names map[string]jsonschema.Property
			if s != nil {
				names = make(map[string]jsonschema.Property, len(s.Properties))
				for _, prop := range s.Properties {
					names[prop.Name] = prop
				}
			}

			parseEncoding := func(name string, prop jsonschema.Property, e ogen.Encoding) (rerr error) {
				locator := e.Common.Locator
				defer func() {
					rerr = p.wrapLocation(p.file(ctx), locator, rerr)
				}()

				encoding := &openapi.Encoding{
					ContentType:   e.ContentType,
					Headers:       map[string]*openapi.Header{},
					Style:         inferParamStyle(openapi.LocationQuery, e.Style),
					Explode:       inferParamExplode(openapi.LocationQuery, e.Explode),
					AllowReserved: e.AllowReserved,
					Pointer:       locator.Pointer(p.file(ctx)),
				}
				encoding.Headers, err = p.parseHeaders(e.Headers, ctx)
				if err != nil {
					return p.wrapField("headers", p.file(ctx), locator, err)
				}
				encodings[name] = encoding

				if err := p.validateParamStyle(&openapi.Parameter{
					Name:          name,
					Schema:        prop.Schema,
					In:            openapi.LocationQuery,
					Style:         encoding.Style,
					Explode:       encoding.Explode,
					Required:      prop.Required,
					AllowReserved: encoding.AllowReserved,
					Pointer:       encoding.Pointer,
				}, p.file(ctx)); err != nil {
					return errors.Wrap(err, "param style")
				}

				return nil
			}

			encodingLoc := locator.Field("encoding")
			for name, e := range m.Encoding {
				prop, ok := names[name]
				if !ok {
					loc := encodingLoc.Key(name)
					err := errors.Errorf("unknown property %q", name)
					return nil, p.wrapLocation(p.file(ctx), loc, err)
				}

				if err := parseEncoding(name, prop, e); err != nil {
					return nil, errors.Wrapf(err, "encoding property %q", name)
				}
			}
		}
	}

	examples := make(map[string]*openapi.Example, len(m.Examples))
	for name, ex := range m.Examples {
		examples[name], err = p.parseExample(ex, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "examples: %q", name)
		}
	}

	// OpenAPI 3.0.3 doc says:
	//
	//   Furthermore, referencing a schema which contains an example,
	//   the example value SHALL override the example provided by the schema.
	//
	// Probably this will be rewritten later.
	// Kept for backward compatibility.
	s.AddExample(m.Example)
	for _, ex := range examples {
		s.AddExample(ex.Value)
	}

	var streaming bool
	{
		const extensionName = "x-ogen-json-streaming"
		if ex, ok := m.Common.Extensions[extensionName]; ok {
			if err := ex.Decode(&streaming); err != nil {
				err := errors.Wrap(err, "unmarshal value")
				return nil, p.wrapField(extensionName, p.file(ctx), locator, err)
			}
		}
	}

	return &openapi.MediaType{
		Schema:             s,
		Example:            json.RawMessage(m.Example),
		Examples:           examples,
		Encoding:           encodings,
		XOgenJSONStreaming: streaming,
		Pointer:            locator.Pointer(p.file(ctx)),
	}, nil
}
