package gen

import (
	"net/http"
	"reflect"
	"sort"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) fix() {
	for _, op := range g.operations {
		g.fixEqualRequests(op)
		g.fixEqualResponses(op)
	}
}

func (g *Generator) fixEqualResponses(op *ir.Operation) {
	if !op.Response.Type.Is(ir.KindInterface) {
		return
	}

	// We can modify contents of operation response.
	// To prevent changes affecting to other operations
	// (in case of referenced responses), we copy the response.
	op.Response = cloneResponse(op.Response)

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

		// TODO: Fix duplicates.
		// g.saveType(alias)
		g.types[alias.Alias.Name] = alias

		if candidate.replaceNoc {
			candidate.response.NoContent = alias
			continue
		}

		candidate.response.Contents[ir.ContentType(candidate.replaceCtype)] = alias
	}
}

func cloneResponse(r *ir.Response) *ir.Response {
	newR := &ir.Response{
		Type:       r.Type,
		StatusCode: map[int]*ir.StatusResponse{},
		Default:    r.Default,
		Spec:       r.Spec,
	}
	for code, statResp := range r.StatusCode {
		newStatResp := &ir.StatusResponse{
			NoContent: statResp.NoContent,
			Contents:  map[ir.ContentType]*ir.Type{},
		}
		for contentType, typ := range statResp.Contents {
			newStatResp.Contents[contentType] = typ
		}
		newR.StatusCode[code] = newStatResp
	}
	return newR
}

func (g *Generator) fixEqualRequests(op *ir.Operation) {
	if op.Request == nil {
		return
	}
	if !op.Request.Type.Is(ir.KindInterface) {
		return
	}

	// We can modify request contents.
	// To prevent changes affecting to other operations
	// (in case of referenced requestBodies), we copy requestBody.
	op.Request = cloneRequest(op.Request)

	type candidate struct {
		renameTo string
		ctype    string
		typ      *ir.Type
	}
	var candidates []candidate

	var contents []string
	for ct := range op.Request.Contents {
		contents = append(contents, string(ct))
	}
	sort.Strings(contents)

	for _, lcontent := range contents {
		lschema := op.Request.Contents[ir.ContentType(lcontent)]
		for _, rcontent := range contents {
			if lcontent == rcontent {
				continue
			}

			rschema := op.Request.Contents[ir.ContentType(rcontent)]
			if reflect.DeepEqual(lschema, rschema) {
				candidates = append(candidates, candidate{
					renameTo: pascal(op.Name, lcontent),
					ctype:    lcontent,
					typ:      lschema,
				})
				candidates = append(candidates, candidate{
					renameTo: pascal(op.Name, rcontent),
					ctype:    rcontent,
					typ:      rschema,
				})
			}
		}
	}

	for _, candidate := range candidates {
		candidate.typ.Unimplement(op.Request.Type)
		alias := ir.Alias(candidate.renameTo, candidate.typ)
		alias.Implement(op.Request.Type)

		// TODO: Fix duplicates.
		// g.saveType(alias)
		g.types[alias.Alias.Name] = alias

		op.Request.Contents[ir.ContentType(candidate.ctype)] = alias
	}
}

func cloneRequest(r *ir.Request) *ir.Request {
	contents := make(map[ir.ContentType]*ir.Type)
	for contentType, typ := range r.Contents {
		contents[contentType] = typ
	}
	return &ir.Request{
		Type:     r.Type,
		Contents: contents,
		Required: r.Required,
		Spec:     r.Spec,
	}
}
