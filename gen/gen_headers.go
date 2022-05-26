package gen

import (
	"net/http"

	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateHeaders(ctx *genctx, name string, headers map[string]*openapi.Header) (_ map[string]*ir.Parameter, err error) {
	if len(headers) == 0 {
		return nil, nil
	}

	result := make(map[string]*ir.Parameter, len(headers))
	for hname, header := range headers {
		ctx := ctx.appendPath(hname)
		if http.CanonicalHeaderKey(hname) == "Content-Type" {
			continue
		}

		result[hname], err = g.generateParameter(ctx, name, header)
		if err != nil {
			if err := g.fail(err); err != nil {
				return nil, errors.Wrap(err, hname)
			}

			continue
		}
	}

	return result, nil
}
