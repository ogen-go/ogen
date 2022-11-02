package gen

import (
	"net/http"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateHeaders(
	ctx *genctx,
	name string,
	headers map[string]*openapi.Header,
) (_ map[string]*ir.Parameter, err error) {
	if len(headers) == 0 {
		return nil, nil
	}

	result := make(map[string]*ir.Parameter, len(headers))
	for hname, header := range headers {
		ctx := ctx.appendPath(hname)
		if http.CanonicalHeaderKey(hname) == "Content-Type" {
			g.log.Warn(
				"Content-Type is described separately and will be ignored in this section.",
				zapPosition(header),
			)
			continue
		}

		result[hname], err = g.generateParameter(ctx, name, header)
		if err != nil {
			if err := g.trySkip(err, "Skipping response header", header); err != nil {
				return nil, err
			}

			delete(result, hname)
			continue
		}
	}

	return result, nil
}
