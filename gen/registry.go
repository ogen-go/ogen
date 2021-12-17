package gen

import (
	"fmt"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) saveIface(t *ir.Type) {
	if !t.Is(ir.KindInterface) {
		panic("unreachable")
	}

	if _, ok := g.interfaces[t.Name]; ok {
		panic(fmt.Sprintf("interface name conflict: %q", t.Name))
	}

	g.interfaces[t.Name] = t
}

func (g *Generator) saveType(t *ir.Type) {
	if !t.Is(ir.KindStruct, ir.KindEnum, ir.KindAlias, ir.KindGeneric, ir.KindSum, ir.KindStream) {
		panic("unreachable")
	}

	if _, ok := g.types[t.Name]; ok && !t.IsGeneric() {
		panic(fmt.Sprintf("schema name conflict: %q", t.Name))
	}

	g.types[t.Name] = t
}

func (g *Generator) saveRef(ref string, t *ir.Type) {
	if !t.Is(ir.KindStruct, ir.KindEnum, ir.KindAlias, ir.KindGeneric, ir.KindSum) {
		panic("unreachable")
	}

	if _, ok := g.refs.types[ref]; ok && !t.IsGeneric() {
		panic(fmt.Sprintf("ref conflict: %q", ref))
	}

	if _, ok := g.types[t.Name]; ok {
		panic(fmt.Sprintf("ref name conflict: %q", t.Name))
	}

	g.refs.types[ref] = t
	g.types[t.Name] = t
}
