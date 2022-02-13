package gen

import (
	"net/http"
	"reflect"
	"sort"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

// deduceDefault implements convenient errors, representing common default
// response as error instead of variant of each response.
func (g *Generator) reduceDefault(ops []*oas.Operation) error {
	if len(ops) < 1 {
		return nil
	}

	// Compare first default response to others.
	first := ops[0]
	if first.Responses == nil || first.Responses["default"] == nil {
		return nil
	}
	d := first.Responses["default"]
	if d.Ref == "" {
		// Not supported.
		return nil
	}
	for _, spec := range ops[1:] {
		if !reflect.DeepEqual(spec.Responses["default"], d) {
			return nil
		}
	}

	ctx := &genctx{
		path:   []string{"x-ogen-reduce-default"},
		global: g.tstorage,
		local:  g.tstorage,
	}

	resp, err := g.responseToIR(ctx, "ErrResp", "reduced default response", d)
	if err != nil {
		return errors.Wrap(err, "default")
	}
	if resp.NoContent != nil || len(resp.Contents) > 1 || resp.Contents[ir.ContentTypeJSON] == nil {
		return errors.Wrap(err, "too complicated to reduce default error")
	}

	g.errType, err = wrapResponseStatusCode(ctx, "", resp)
	if err != nil {
		return errors.Wrap(err, "wrap default response with status code struct")
	}

	return nil
}

// Example:
// ...
// responses:
//   200:
//     contents:
//       application/json:
//         ref: #/components/schemas/Foo
//   202:
//     contents:
//       application/json:
//         ref: #/components/schemas/Foo
//
// This response refers to the same schema for different
// status codes, and it will cause a collision:
//
// func encodeResponse(resp FooResponse) {
//     switch resp.(type) {
//	   case *Foo:
//     case *Foo:
//     }
// }
//
// To prevent collision we wrap referenced schema with aliases
// and use them instead.
//
// type FooResponseOK Foo
// func(*FooResponseOK) FooResponse() {}
//
// type FooResponseAccepted Foo
// func(*FooResponseOK) FooResponse() {}
//
// Referring to the same schema in different content types
// also can cause a collision.
func reduceEqualResponses(ctx *genctx, op *ir.Operation) {
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

		replaceNoc bool
		replaceCT  string
		response   *ir.StatusResponse
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
							renameTo:  pascal(op.Name, lct, http.StatusText(lcode)),
							typ:       lschema,
							replaceCT: lct,
							response:  lresp,
						})
						candidates = append(candidates, candidate{
							renameTo:  pascal(op.Name, rct, http.StatusText(rcode)),
							typ:       rschema,
							replaceCT: rct,
							response:  rresp,
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
		ctx.local.types[alias.Name] = alias

		if candidate.replaceNoc {
			candidate.response.NoContent = alias
			continue
		}

		candidate.response.Contents[ir.ContentType(candidate.replaceCT)] = alias
	}
}

func cloneResponse(r *ir.Response) *ir.Response {
	newR := &ir.Response{
		Type:       r.Type,
		StatusCode: map[int]*ir.StatusResponse{},
		Default:    r.Default,
	}
	for code, statResp := range r.StatusCode {
		newStatResp := &ir.StatusResponse{
			NoContent: statResp.NoContent,
			Contents:  map[ir.ContentType]*ir.Type{},
		}
		for contentType, t := range statResp.Contents {
			newStatResp.Contents[contentType] = t
		}
		newR.StatusCode[code] = newStatResp
	}
	return newR
}

func reduceEqualRequests(ctx *genctx, op *ir.Operation) {
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
		t        *ir.Type
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
					t:        lschema,
				})
				candidates = append(candidates, candidate{
					renameTo: pascal(op.Name, rcontent),
					ctype:    rcontent,
					t:        rschema,
				})
			}
		}
	}

	for _, candidate := range candidates {
		candidate.t.Unimplement(op.Request.Type)
		alias := ir.Alias(candidate.renameTo, candidate.t)
		alias.Implement(op.Request.Type)

		// TODO: Fix duplicates.
		// g.saveType(alias)
		ctx.local.types[alias.Name] = alias

		op.Request.Contents[ir.ContentType(candidate.ctype)] = alias
	}
}

func cloneRequest(r *ir.Request) *ir.Request {
	contents := make(map[ir.ContentType]*ir.Type)
	for contentType, t := range r.Contents {
		contents[contentType] = t
	}
	return &ir.Request{
		Type:     r.Type,
		Contents: contents,
		Spec:     r.Spec,
	}
}
