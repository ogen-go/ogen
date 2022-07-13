package gen

import (
	"reflect"
	"sort"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
)

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
// func(*FooResponseAccepted) FooResponse() {}
//
// Referring to the same schema in different content types
// also can cause a collision and it will be fixed in the same way.
func fixEqualResponses(ctx *genctx, op *ir.Operation) error {
	if !op.Responses.Type.Is(ir.KindInterface) {
		return nil
	}

	// We can modify contents of operation response.
	// To prevent changes affecting to other operations
	// (in case of referenced responses), we copy the response.
	op.Responses = cloneResponse(op.Responses)

	var statusCodes []int
	for code := range op.Responses.StatusCode {
		statusCodes = append(statusCodes, code)
	}
	sort.Ints(statusCodes)

	type candidate struct {
		renameTo string
		encoding ir.Encoding
		typ      *ir.Type

		replaceNoc bool
		replaceCT  string
		response   *ir.Response
	}

	var candidates []candidate
	for i := 0; i < len(statusCodes); i++ {
		lcode := statusCodes[i]
		for j := i; j < len(statusCodes); j++ {
			rcode := statusCodes[j]
			lresp, rresp := op.Responses.StatusCode[lcode], op.Responses.StatusCode[rcode]
			if (lresp.NoContent != nil && rresp.NoContent != nil) && lcode != rcode {
				if reflect.DeepEqual(lresp.NoContent, rresp.NoContent) {
					lname, err := pascal(op.Name, statusText(lcode))
					if err != nil {
						return errors.Wrap(err, "lname")
					}
					rname, err := pascal(op.Name, statusText(rcode))
					if err != nil {
						return errors.Wrap(err, "rname")
					}

					candidates = append(candidates, candidate{
						renameTo:   lname,
						typ:        lresp.NoContent,
						replaceNoc: true,
						response:   lresp,
					}, candidate{
						renameTo:   rname,
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
					lmedia, rmedia := lresp.Contents[ir.ContentType(lct)], rresp.Contents[ir.ContentType(rct)]
					ltype, rtype := lmedia.Type, rmedia.Type
					if reflect.DeepEqual(ltype, rtype) {
						lname, err := pascal(op.Name, lct, statusText(lcode))
						if err != nil {
							return errors.Wrap(err, "lname")
						}
						rname, err := pascal(op.Name, rct, statusText(rcode))
						if err != nil {
							return errors.Wrap(err, "rname")
						}

						candidates = append(candidates, candidate{
							renameTo:  lname,
							encoding:  lmedia.Encoding,
							typ:       ltype,
							replaceCT: lct,
							response:  lresp,
						}, candidate{
							renameTo:  rname,
							encoding:  rmedia.Encoding,
							typ:       rtype,
							replaceCT: rct,
							response:  rresp,
						})
					}
				}
			}
		}
	}

	for _, candidate := range candidates {
		candidate.typ.Unimplement(op.Responses.Type)
		alias := ir.Alias(candidate.renameTo, candidate.typ)
		alias.Implement(op.Responses.Type)

		// TODO: Fix duplicates.
		// g.saveType(alias)
		ctx.local.types[alias.Name] = alias

		if candidate.replaceNoc {
			candidate.response.NoContent = alias
			continue
		}

		candidate.response.Contents[ir.ContentType(candidate.replaceCT)] = ir.Media{
			Encoding: candidate.encoding,
			Type:     alias,
		}
	}

	return nil
}

func cloneResponse(r *ir.Responses) *ir.Responses {
	newR := &ir.Responses{
		Type:       r.Type,
		StatusCode: map[int]*ir.Response{},
		Default:    r.Default,
	}
	for code, statResp := range r.StatusCode {
		newStatResp := &ir.Response{
			NoContent:      statResp.NoContent,
			Contents:       map[ir.ContentType]ir.Media{},
			Headers:        statResp.Headers,
			WithStatusCode: statResp.WithStatusCode,
			WithHeaders:    statResp.WithHeaders,
		}
		for contentType, t := range statResp.Contents {
			newStatResp.Contents[contentType] = t
		}
		newR.StatusCode[code] = newStatResp
	}
	return newR
}

func fixEqualRequests(ctx *genctx, op *ir.Operation) error {
	if op.Request == nil {
		return nil
	}
	if !op.Request.Type.Is(ir.KindInterface) {
		return nil
	}

	// We can modify request contents.
	// To prevent changes affecting to other operations
	// (in case of referenced requestBodies), we copy requestBody.
	op.Request = cloneRequest(op.Request)

	type candidate struct {
		renameTo string
		ctype    string
		encoding ir.Encoding
		t        *ir.Type
	}
	var candidates []candidate

	var contents []string
	for ct := range op.Request.Contents {
		contents = append(contents, string(ct))
	}
	sort.Strings(contents)

	for _, lcontent := range contents {
		lmedia := op.Request.Contents[ir.ContentType(lcontent)]
		ltype := lmedia.Type
		for _, rcontent := range contents {
			if lcontent == rcontent {
				continue
			}

			rmedia := op.Request.Contents[ir.ContentType(rcontent)]
			rtype := rmedia.Type
			if reflect.DeepEqual(ltype, rtype) {
				lname, err := pascal(op.Name, lcontent)
				if err != nil {
					return errors.Wrap(err, "lname")
				}
				rname, err := pascal(op.Name, rcontent)
				if err != nil {
					return errors.Wrap(err, "rname")
				}
				candidates = append(candidates, candidate{
					renameTo: lname,
					ctype:    lcontent,
					encoding: lmedia.Encoding,
					t:        ltype,
				}, candidate{
					renameTo: rname,
					ctype:    rcontent,
					encoding: rmedia.Encoding,
					t:        rtype,
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

		op.Request.Contents[ir.ContentType(candidate.ctype)] = ir.Media{
			Encoding: candidate.encoding,
			Type:     alias,
		}
	}

	return nil
}

func cloneRequest(r *ir.Request) *ir.Request {
	contents := make(map[ir.ContentType]ir.Media)
	for contentType, media := range r.Contents {
		contents[contentType] = media
	}
	return &ir.Request{
		Type:     r.Type,
		Contents: contents,
		Spec:     r.Spec,
	}
}
