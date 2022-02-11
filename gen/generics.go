package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) wrapGenerics() {
	for _, t := range g.types {
		if t.Is(ir.KindStruct) || (t.Is(ir.KindMap) && len(t.Fields) > 0) {
			g.boxStructFields(t)
		}
	}

	wrapType := func(optional bool, typ *ir.Type) *ir.Type {
		if typ == nil {
			return nil
		}
		if typ.Is(ir.KindStream) {
			// Do not wrap io.Reader requests.
			return typ
		}

		v := ir.GenericVariant{
			Nullable: typ.Schema != nil && typ.Schema.Nullable,
			Optional: optional,
		}
		if v.Any() {
			return g.boxType(v, typ)
		}
		return typ
	}
	wrapContents := func(optional bool, contents map[ir.ContentType]*ir.Type) {
		for contentType, typ := range contents {
			contents[contentType] = wrapType(optional, typ)
		}
	}

	for _, op := range g.operations {
		for _, param := range op.Params {
			v := ir.GenericVariant{
				Nullable: param.Spec.Schema.Nullable,
				Optional: !param.Spec.Required,
			}

			if v.Any() {
				param.Type = g.boxType(v, param.Type)
			}
		}

		if req := op.Request; req != nil {
			optional := req.Spec != nil && !req.Spec.Required
			if !req.Type.IsInterface() {
				req.Type = wrapType(optional, req.Type)
			}
			wrapContents(optional, req.Contents)
		}

		wrapStatusResponse := func(r *ir.StatusResponse) {
			if r == nil {
				return
			}
			r.NoContent = wrapType(false, r.NoContent)
			wrapContents(false, r.Contents)
		}
		if resp := op.Response; resp != nil {
			if !resp.Type.IsInterface() {
				resp.Type = wrapType(false, resp.Type)
			}
			wrapStatusResponse(resp.Default)
			for _, r := range resp.StatusCode {
				wrapStatusResponse(r)
			}
		}
	}
}

func (g *Generator) boxStructFields(s *ir.Type) {
	for _, field := range s.Fields {
		if field.Spec == nil {
			continue
		}

		v := ir.GenericVariant{
			Nullable: field.Spec.Schema.Nullable,
			Optional: !field.Spec.Required,
		}

		field.Type = func(t *ir.Type) *ir.Type {
			if s.RecursiveTo(t) {
				switch {
				case v.OnlyOptional():
					return ir.Pointer(t, ir.NilOptional)
				case v.OnlyNullable():
					return ir.Pointer(t, ir.NilNull)
				case v.NullableOptional():
					return ir.Pointer(g.boxType(ir.GenericVariant{
						Optional: true,
					}, t), ir.NilNull)
				default:
					// Required.
					panic(fmt.Sprintf("recursion: %s.%s", s.Name, field.Name))
				}
			}
			if v.Any() {
				return g.boxType(v, t)
			}
			return t
		}(field.Type)
	}
}

func (g *Generator) boxType(v ir.GenericVariant, t *ir.Type) *ir.Type {
	if t.IsAny() {
		// Do not wrap Any.
		return t
	}
	if t.IsArray() || t.Primitive == ir.ByteSlice {
		// Using special case for array nil value if possible.
		switch {
		case v.OnlyOptional():
			t.NilSemantic = ir.NilOptional
		case v.OnlyNullable():
			t.NilSemantic = ir.NilNull
		default:
			t = ir.Generic(genericPostfix(t),
				t, v,
			)
			g.saveType(t)
		}

		return t
	}

	if t.CanGeneric() {
		t = ir.Generic(genericPostfix(t), t, v)
		g.saveType(t)
		return t
	}

	switch {
	case v.OnlyOptional():
		return t.Pointer(ir.NilOptional)
	case v.OnlyNullable():
		return t.Pointer(ir.NilNull)
	default:
		t = ir.Generic(genericPostfix(t),
			t.Pointer(ir.NilNull), ir.GenericVariant{Optional: true},
		)
		g.saveType(t)
		return t
	}
}

func genericPostfix(t *ir.Type) string {
	name := t.NamePostfix()
	if idx := strings.Index(name, "."); idx > 0 {
		name = name[idx+1:]
	}
	return pascal(name)
}
