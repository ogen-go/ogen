package gen

import (
	"fmt"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) saveIface(typ *ir.Type) {
	if !typ.Is(ir.KindInterface) {
		panic("unreachable")
	}

	if _, ok := g.interfaces[typ.Name]; ok {
		panic(fmt.Sprintf("interface name conflict: '%s'", typ.Name))
	}

	g.interfaces[typ.Name] = typ
}

func (g *Generator) saveType(typ *ir.Type) {
	if !typ.Is(ir.KindStruct, ir.KindEnum, ir.KindAlias, ir.KindGeneric, ir.KindSum) {
		panic("unreachable")
	}

	if _, ok := g.types[typ.Name]; ok && !typ.IsGeneric() {
		panic(fmt.Sprintf("schema name conflict: '%s'", typ.Name))
	}

	g.types[typ.Name] = typ
}

func (g *Generator) saveRef(ref string, typ *ir.Type) {
	if !typ.Is(ir.KindStruct, ir.KindEnum, ir.KindAlias, ir.KindGeneric, ir.KindSum) {
		panic("unreachable")
	}

	if _, ok := g.refs.schemas[ref]; ok && !typ.IsGeneric() {
		panic(fmt.Sprintf("ref conflict: '%s'", ref))
	}

	if _, ok := g.types[typ.Name]; ok {
		panic(fmt.Sprintf("ref name conflict: '%s'", typ.Name))
	}

	g.refs.schemas[ref] = typ
	g.types[typ.Name] = typ
}
