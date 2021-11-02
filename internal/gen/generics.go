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
	for _, field := range s.Fields {
		if field.Spec == nil {
			continue
		}

		v := ir.GenericVariant{
			Nullable: field.Spec.Schema.Nullable,
			Optional: !field.Spec.Required,
		}

		field.Type = func(typ *ir.Type) *ir.Type {
			if s.RecursiveTo(typ) {
				switch {
				case v.Nullable && !v.Optional: // nullable
					return ir.Pointer(typ, ir.NilNull)
				case !v.Nullable && v.Optional: // optional
					return ir.Pointer(typ, ir.NilOptional)
				case v.Nullable && v.Optional: // nullable & optional
					return ir.Pointer(g.boxType(ir.GenericVariant{
						Optional: true,
					}, typ), ir.NilNull)
				case !v.Nullable && !v.Optional: // required
					panic(fmt.Sprintf("recursion: %s.%s", s.Name, field.Name))
				}
			} else if v.Any() {
				return g.boxType(v, typ)
			}
			return typ
		}(field.Type)
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
