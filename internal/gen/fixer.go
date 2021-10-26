package gen

import (
	"net/http"
	"reflect"
	"sort"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) fix() {
	for _, op := range g.operations {
		g.fixEqualResponses(op)
	}
}

func (g *Generator) fixEqualResponses(op *ir.Operation) {
	if !op.Response.Type.Is(ir.KindInterface) {
		return
	}

	var statusCodes []int
	for code := range op.Response.StatusCode {
		statusCodes = append(statusCodes, code)
	}
	sort.Ints(statusCodes)

	type candidate struct {
		renameTo string
		typ      *ir.Type

		replaceNoc   bool
		replaceCtype string
		response     *ir.StatusResponse
	}

	var candidates []candidate
	for i := 0; i < len(statusCodes); i++ {
		lcode := statusCodes[i]
		for j := i; j < len(statusCodes); j++ {
			rcode := statusCodes[j]
			lresp, rresp := op.Response.StatusCode[lcode], op.Response.StatusCode[rcode]
			if (lresp.NoContent != nil && rresp.NoContent != nil) && lcode != rcode {
				if reflect.DeepEqual(lresp.NoContent, rresp.NoContent) {
					candidates = append(candidates, candidate{
						renameTo:   pascal(op.Name, http.StatusText(lcode)),
						typ:        lresp.NoContent,
						replaceNoc: true,
						response:   lresp,
					})
					candidates = append(candidates, candidate{
						renameTo:   pascal(op.Name, http.StatusText(rcode)),
						typ:        rresp.NoContent,
						replaceNoc: true,
						response:   rresp,
					})
					continue
				}
			}

			var (
				lcontents []string
				rcontents []string
			)
			for ct := range lresp.Contents {
				lcontents = append(lcontents, string(ct))
			}
			for ct := range rresp.Contents {
				rcontents = append(rcontents, string(ct))
			}
			sort.Strings(lcontents)
			sort.Strings(rcontents)
			for _, lct := range lcontents {
				for _, rct := range rcontents {
					if lcode == rcode && lct == rct {
						continue
					}
					lschema, rschema := lresp.Contents[ir.ContentType(lct)], rresp.Contents[ir.ContentType(rct)]
					if reflect.DeepEqual(lschema, rschema) {
						candidates = append(candidates, candidate{
							renameTo:     pascal(op.Name, lct, http.StatusText(lcode)),
							typ:          lschema,
							replaceCtype: lct,
							response:     lresp,
						})
						candidates = append(candidates, candidate{
							renameTo:     pascal(op.Name, rct, http.StatusText(rcode)),
							typ:          rschema,
							replaceCtype: rct,
							response:     rresp,
						})
					}
				}
			}
		}
	}

	for _, candidate := range candidates {
		candidate.typ.Unimplement(op.Response.Type)
		alias := ir.Alias(candidate.renameTo, candidate.typ)
		alias.Implement(op.Response.Type)
		g.saveType(alias)
		if candidate.replaceNoc {
			candidate.response.NoContent = alias
			continue
		}

		candidate.response.Contents[ir.ContentType(candidate.replaceCtype)] = alias
	}
}
