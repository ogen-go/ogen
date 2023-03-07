package gen

import (
	"reflect"

	"github.com/go-faster/errors"
	"golang.org/x/exp/maps"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
)

// Example:
//
//	responses:
//		200:
//		  contents:
//		    application/json:
//		      ref: #/components/schemas/Foo
//		202:
//		  contents:
//		    application/json:
//		      ref: #/components/schemas/Foo
//
// This response refers to the same schema for different
// status codes, and it will cause a collision:
//
//	func encodeResponse(resp FooResponse) {
//	    switch resp.(type) {
//		case *Foo:
//	    case *Foo:
//	    }
//	}
//
// To prevent collision we wrap referenced schema with aliases
// and use them instead.
//
//	type FooResponseOK Foo
//	func(*FooResponseOK) FooResponse() {}
//
//	type FooResponseAccepted Foo
//	func(*FooResponseAccepted) FooResponse() {}
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

	statusCodes := xmaps.SortedKeys(op.Responses.StatusCode)
	type candidate struct {
		renameTo      string
		encoding      ir.Encoding
		JSONStreaming bool
		typ           *ir.Type

		replaceNoc bool
		replaceCT  ir.ContentType
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
				lcontents = xmaps.SortedKeys(lresp.Contents)
				rcontents = xmaps.SortedKeys(rresp.Contents)

				// Add `Content-Type` to response name only if needed.
				trySkipCT = func(s ir.ContentType, contents []ir.ContentType) string {
					if len(contents) > 1 {
						return string(s)
					}
					return ""
				}
			)
			for _, lct := range lcontents {
				for _, rct := range rcontents {
					if lcode == rcode && lct == rct {
						continue
					}
					lmedia, rmedia := lresp.Contents[lct], rresp.Contents[rct]
					ltype, rtype := lmedia.Type, rmedia.Type
					if reflect.DeepEqual(ltype, rtype) {
						lname, err := pascal(op.Name, trySkipCT(lct, lcontents), statusText(lcode))
						if err != nil {
							return errors.Wrap(err, "lname")
						}

						rname, err := pascal(op.Name, trySkipCT(rct, rcontents), statusText(rcode))
						if err != nil {
							return errors.Wrap(err, "rname")
						}

						candidates = append(candidates, candidate{
							renameTo:      lname,
							encoding:      lmedia.Encoding,
							JSONStreaming: lmedia.JSONStreaming,
							typ:           ltype,
							replaceCT:     lct,
							response:      lresp,
						}, candidate{
							renameTo:      rname,
							encoding:      rmedia.Encoding,
							JSONStreaming: rmedia.JSONStreaming,
							typ:           rtype,
							replaceCT:     rct,
							response:      rresp,
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

		candidate.response.Contents[candidate.replaceCT] = ir.Media{
			Encoding:      candidate.encoding,
			Type:          alias,
			JSONStreaming: candidate.JSONStreaming,
		}
	}

	return nil
}

func cloneResponse(r *ir.Responses) *ir.Responses {
	cloneResponse := func(r *ir.Response) *ir.Response {
		if r == nil {
			return nil
		}

		return &ir.Response{
			NoContent:      r.NoContent,
			Contents:       maps.Clone(r.Contents),
			Headers:        r.Headers,
			WithStatusCode: r.WithStatusCode,
			WithHeaders:    r.WithHeaders,
		}
	}

	newR := &ir.Responses{
		Type:       r.Type,
		StatusCode: make(map[int]*ir.Response, len(r.StatusCode)),
		Default:    cloneResponse(r.Default),
	}
	for code, resp := range r.StatusCode {
		newR.StatusCode[code] = cloneResponse(resp)
	}
	for idx, resp := range r.Pattern {
		newR.Pattern[idx] = cloneResponse(resp)
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
		renameTo      string
		ctype         ir.ContentType
		encoding      ir.Encoding
		JSONStreaming bool
		t             *ir.Type
	}
	var (
		candidates []candidate
		contents   = xmaps.SortedKeys(op.Request.Contents)
	)

	for _, lcontent := range contents {
		lmedia := op.Request.Contents[lcontent]
		ltype := lmedia.Type
		for _, rcontent := range contents {
			if lcontent == rcontent {
				continue
			}

			rmedia := op.Request.Contents[rcontent]
			rtype := rmedia.Type
			if reflect.DeepEqual(ltype, rtype) {
				lname, err := pascal(op.Name, string(lcontent))
				if err != nil {
					return errors.Wrap(err, "lname")
				}
				rname, err := pascal(op.Name, string(rcontent))
				if err != nil {
					return errors.Wrap(err, "rname")
				}
				candidates = append(candidates, candidate{
					renameTo:      lname,
					ctype:         lcontent,
					encoding:      lmedia.Encoding,
					JSONStreaming: lmedia.JSONStreaming,
					t:             ltype,
				}, candidate{
					renameTo:      rname,
					ctype:         rcontent,
					encoding:      rmedia.Encoding,
					JSONStreaming: rmedia.JSONStreaming,
					t:             rtype,
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

		op.Request.Contents[candidate.ctype] = ir.Media{
			Encoding:      candidate.encoding,
			Type:          alias,
			JSONStreaming: candidate.JSONStreaming,
		}
	}

	return nil
}

func cloneRequest(r *ir.Request) *ir.Request {
	return &ir.Request{
		Type:      r.Type,
		EmptyBody: r.EmptyBody,
		Contents:  maps.Clone(r.Contents),
		Spec:      r.Spec,
	}
}
