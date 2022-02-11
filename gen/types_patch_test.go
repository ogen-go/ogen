package gen

import (
	"testing"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/stretchr/testify/require"
)

func TestRequestPatch(t *testing.T) {
	t.Run("SingleType", func(t *testing.T) {
		originalType := ir.Primitive(ir.Bool, nil)
		body := &ir.Request{
			Type: originalType,
			Contents: map[ir.ContentType]*ir.Type{
				ir.ContentTypeJSON: originalType,
			},
		}
		expectBody := &ir.Request{
			Type: ir.Pointer(originalType, ir.NilOptional),
			Contents: map[ir.ContentType]*ir.Type{
				ir.ContentTypeJSON: ir.Pointer(originalType, ir.NilOptional),
			},
		}

		calls := 0
		patchRequestTypes(body, func(_ string, inspectingT *ir.Type) *ir.Type {
			require.Equal(t, originalType, inspectingT)
			calls++
			return ir.Pointer(inspectingT, ir.NilOptional)
		})

		require.Equal(t, 1, calls)
		require.Equal(t, expectBody, body)
	})

	t.Run("MultiType", func(t *testing.T) {
		t.Skip()
		var (
			iface            = ir.Interface("FooReq")
			jsonT            = ir.Primitive(ir.Bool, nil)
			streamT *ir.Type = nil

			expectJsonT   = ir.Pointer(jsonT, ir.NilOptional)
			expectStreamT = ir.Pointer(streamT, ir.NilOptional)
		)

		body := &ir.Request{
			Type: iface,
			Contents: map[ir.ContentType]*ir.Type{
				ir.ContentTypeJSON:        jsonT,
				ir.ContentTypeOctetStream: streamT,
			},
		}
		expectBody := &ir.Request{
			Type: iface,
			Contents: map[ir.ContentType]*ir.Type{
				ir.ContentTypeJSON:        expectJsonT,
				ir.ContentTypeOctetStream: expectStreamT,
			},
		}

		visited := map[*ir.Type]struct{}{}
		patchRequestTypes(body, func(_ string, inspectingT *ir.Type) *ir.Type {
			visited[inspectingT] = struct{}{}
			return ir.Pointer(inspectingT, ir.NilOptional)
		})

		require.Equal(t, 2, len(visited))
		require.Equal(t, expectBody, body)
	})
}

func TestResponsePatch(t *testing.T) {
	t.Run("SingleType", func(t *testing.T) {
		typ := ir.Primitive(ir.Bool, nil)
		resp := &ir.Response{
			Type: typ,
			StatusCode: map[int]*ir.StatusResponse{
				200: {
					Contents: map[ir.ContentType]*ir.Type{
						ir.ContentTypeJSON: typ,
					},
				},
			},
		}

		expectTyp := ir.Pointer(typ, ir.NilOptional)
		expectResp := &ir.Response{
			Type: expectTyp,
			StatusCode: map[int]*ir.StatusResponse{
				200: {
					Contents: map[ir.ContentType]*ir.Type{
						ir.ContentTypeJSON: expectTyp,
					},
				},
			},
		}

		calls := 0
		patchResponseTypes(resp, func(_ string, t *ir.Type) *ir.Type {
			calls++
			return ir.Pointer(t, ir.NilOptional)
		})

		require.Equal(t, 1, calls)
		require.Equal(t, expectResp, resp)
	})

	t.Run("MultiTypeWithDefault", func(t *testing.T) {
		var (
			typ      = ir.Interface("FooRes")
			okT      = ir.Primitive(ir.Bool, nil)
			errT     = ir.Primitive(ir.Bool, nil)
			wrappedT = &ir.Type{
				Kind: ir.KindStruct,
				Fields: []*ir.Field{
					{
						Name: "StatusCode",
						Type: ir.Primitive(ir.Int, nil),
					},
					{
						Name: "Response",
						Type: errT,
					},
				},
			}

			expectOkT      = ir.Pointer(okT, ir.NilOptional)
			expectWrappedT = &ir.Type{
				Kind: ir.KindStruct,
				Fields: []*ir.Field{
					{
						Name: "StatusCode",
						Type: ir.Primitive(ir.Int, nil),
					},
					{
						Name: "Response",
						Type: ir.Pointer(errT, ir.NilOptional),
					},
				},
			}
		)

		resp := &ir.Response{
			Type: typ,
			StatusCode: map[int]*ir.StatusResponse{
				200: {
					Contents: map[ir.ContentType]*ir.Type{
						ir.ContentTypeJSON: okT,
					},
				},
			},
			Default: &ir.StatusResponse{
				Wrapped: true,
				Contents: map[ir.ContentType]*ir.Type{
					ir.ContentTypeJSON: wrappedT,
				},
			},
		}

		expectResp := &ir.Response{
			Type: typ,
			StatusCode: map[int]*ir.StatusResponse{
				200: {
					Contents: map[ir.ContentType]*ir.Type{
						ir.ContentTypeJSON: expectOkT,
					},
				},
			},
			Default: &ir.StatusResponse{
				Wrapped: true,
				Contents: map[ir.ContentType]*ir.Type{
					ir.ContentTypeJSON: expectWrappedT,
				},
			},
		}

		visited := map[*ir.Type]struct{}{}
		patchResponseTypes(resp, func(_ string, t *ir.Type) *ir.Type {
			visited[t] = struct{}{}
			return ir.Pointer(t, ir.NilOptional)
		})

		require.Equal(t, 2, len(visited))
		require.Equal(t, expectResp, resp)
	})
}
