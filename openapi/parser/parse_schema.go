package parser

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
)

func (p *parser) parseSchema(schema *ogen.Schema, ctx *resolveCtx) (_ *jsonschema.Schema, rerr error) {
	if schema != nil {
		defer func() {
			rerr = p.wrapLocation(schema, rerr)
		}()
	}
	s := schema.ToJSONSchema()
	if loc := ctx.lastLoc(); s != nil && s.Ref != "" && loc != "" {
		base, err := url.Parse(loc)
		if err != nil {
			return nil, errors.Wrap(err, "parse base")
		}

		ref, err := url.Parse(s.Ref)
		if err != nil {
			return nil, errors.Wrap(err, "parse ref")
		}

		s.Ref = strings.TrimPrefix(base.ResolveReference(ref).String(), "/")
	}
	return p.schemaParser.Parse(s)
}
