package gen

import (
	"net/http"
	"reflect"
	"sort"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) fix() {
	for _, m := range g.methods {
		g.fixEqualResponses(m)
	}
}

func (g *Generator) fixEqualResponses(m *ir.Method) {
	if !m.Response.Type.Is(ir.KindInterface) {
		return
	}

	var statusCodes []int
	for code := range m.Response.StatusCode {
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
			lresp, rresp := m.Response.StatusCode[lcode], m.Response.StatusCode[rcode]
			if (lresp.NoContent != nil && rresp.NoContent != nil) && lcode != rcode {
				if reflect.DeepEqual(lresp.NoContent, rresp.NoContent) {
					candidates = append(candidates, candidate{
						renameTo:   pascal(m.Name, http.StatusText(lcode)),
						typ:        lresp.NoContent,
						replaceNoc: true,
						response:   lresp,
					})
					candidates = append(candidates, candidate{
						renameTo:   pascal(m.Name, http.StatusText(rcode)),
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
				lcontents = append(lcontents, ct)
			}
			for ct := range rresp.Contents {
				rcontents = append(rcontents, ct)
			}
			sort.Strings(lcontents)
			sort.Strings(rcontents)
			for _, lct := range lcontents {
				for _, rct := range rcontents {
					if lcode == rcode && lct == rct {
						continue
					}
					lschema, rschema := lresp.Contents[lct], rresp.Contents[rct]
					if reflect.DeepEqual(lschema, rschema) {
						candidates = append(candidates, candidate{
							renameTo:     pascal(m.Name, lct, http.StatusText(lcode)),
							typ:          lschema,
							replaceCtype: lct,
							response:     lresp,
						})
						candidates = append(candidates, candidate{
							renameTo:     pascal(m.Name, rct, http.StatusText(rcode)),
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
		candidate.typ.Unimplement(m.Response.Type)
		alias := ir.Alias(candidate.renameTo, candidate.typ)
		alias.Implement(m.Response.Type)
		g.types[alias.Name] = alias
		if candidate.replaceNoc {
			candidate.response.NoContent = alias
			continue
		}

		candidate.response.Contents[candidate.replaceCtype] = alias
	}
}
