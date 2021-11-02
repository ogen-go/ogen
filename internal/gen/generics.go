package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) wrapGenerics() {
	for _, typ := range g.types {
		if typ.Is(ir.KindStruct) {
			g.boxStructFields(typ)
		}
	}
}

func (g *Generator) boxStructFields(s *ir.Type) {
	for i := range s.Fields {
		field := s.Fields[i]
		typ := field.Type
		if field.Spec == nil {
			continue
		}

		v := ir.GenericVariant{
			Nullable: field.Spec.Schema.Nullable,
			Optional: !field.Spec.Required,
		}

		if s.RecursiveTo(typ) {
			switch {
			case v.Nullable && !v.Optional: // nullable
				typ = ir.Pointer(typ, ir.NilNull)
			case !v.Nullable && v.Optional: // optional
				typ = ir.Pointer(typ, ir.NilOptional)
			case v.Nullable && v.Optional: // nullable & optional
				typ = ir.Pointer(g.boxType(ir.GenericVariant{
					Optional: true,
				}, typ), ir.NilNull)
			case !v.Nullable && !v.Optional: // required
				panic(fmt.Sprintf("recursion: %s.%s", s.Name, field.Name))
			}
		} else if v.Any() {
			typ = g.boxType(v, typ)
		}

		field.Type = typ
		s.Fields[i] = field
	}
}

func (g *Generator) boxType(v ir.GenericVariant, typ *ir.Type) *ir.Type {
	if typ.IsArray() {
		// Using special case for array nil value if possible.
		switch {
		case v.OnlyOptional():
			typ.NilSemantic = ir.NilOptional
		case v.OnlyNullable():
			typ.NilSemantic = ir.NilNull
		default:
			typ = ir.Generic(genericPostfix(typ.Go()),
				typ, v,
			)
			g.saveType(typ)
		}

		return typ
	}

	if typ.CanGeneric() {
		typ = ir.Generic(genericPostfix(typ.Go()), typ, v)
		g.saveType(typ)
		return typ
	}

	switch {
	case v.OnlyOptional():
		return typ.Pointer(ir.NilOptional)
	case v.OnlyNullable():
		return typ.Pointer(ir.NilNull)
	default:
		typ = ir.Generic(genericPostfix(typ.Go()),
			typ.Pointer(ir.NilNull), ir.GenericVariant{Optional: true},
		)
		g.saveType(typ)
		return typ
	}
}

func genericPostfix(name string) string {
	if idx := strings.Index(name, "."); idx > 0 {
		name = name[idx+1:]
	}
	return pascal(name)
}
