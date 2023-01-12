package gen

import (
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
		if vetHeaderParameterName(g.log, hname, header, "Content-Type") {
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
