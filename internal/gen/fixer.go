package gen

import (
	"net/http"
	"sort"

	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) fix() {
	for _, m := range g.methods {
		g.fixEqualResponses(m)
	}
}

func (g *Generator) fixEqualResponses(m *ast.Method) {
	iface, ok := m.ResponseType.(*ast.Interface)
	if !ok {
		return
	}

	var statusCodes []int
	for code := range m.Responses.StatusCode {
		statusCodes = append(statusCodes, code)
	}
	sort.Ints(statusCodes)

	type candidate struct {
		renameTo string
		schema   *ast.Schema

		replaceNoc   bool
		replaceCtype string
		response     *ast.Response
	}

	var candidates []candidate
	for i := 0; i < len(statusCodes); i++ {
		lcode := statusCodes[i]
		for j := i; j < len(statusCodes); j++ {
			rcode := statusCodes[j]
			lresp, rresp := m.Responses.StatusCode[lcode], m.Responses.StatusCode[rcode]
			if (lresp.NoContent != nil && rresp.NoContent != nil) && lcode != rcode {
				if lresp.NoContent.Equal(rresp.NoContent) {
					candidates = append(candidates, candidate{
						renameTo:   pascal(m.Name, http.StatusText(lcode)),
						schema:     lresp.NoContent,
						replaceNoc: true,
						response:   lresp,
					})
					candidates = append(candidates, candidate{
						renameTo:   pascal(m.Name, http.StatusText(rcode)),
						schema:     rresp.NoContent,
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
					if lschema.Equal(rschema) {
						candidates = append(candidates, candidate{
							renameTo:     pascal(m.Name, lct, http.StatusText(lcode)),
							schema:       lschema,
							replaceCtype: lct,
							response:     lresp,
						})
						candidates = append(candidates, candidate{
							renameTo:     pascal(m.Name, rct, http.StatusText(rcode)),
							schema:       rschema,
							replaceCtype: rct,
							response:     rresp,
						})
					}
				}
			}
		}
	}

	for _, candidate := range candidates {
		candidate.schema.Unimplement(iface)
		alias := ast.Alias(candidate.renameTo, candidate.schema)
		alias.Implement(iface)
		g.schemas[alias.Name] = alias
		if candidate.replaceNoc {
			candidate.response.NoContent = alias
			continue
		}

		candidate.response.Contents[candidate.replaceCtype] = alias
	}
}
