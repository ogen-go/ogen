package gen

import (
	"fmt"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) saveIface(typ *ir.Type) {
	if !typ.Is(ir.KindInterface) {
		panic("unreachable")
	}

	if _, ok := g.interfaces[typ.Interface.Name]; ok {
		panic(fmt.Sprintf("interface name conflict: %q", typ.Interface.Name))
	}

	g.interfaces[typ.Interface.Name] = typ.Interface
}

func (g *Generator) saveType(typ *ir.Type) {
	name, ok := typ.Name()
	if !ok {
		panic("unreachable")
	}

	if _, ok := g.types[name]; ok && !typ.IsGeneric() {
		panic(fmt.Sprintf("schema name conflict: %q", name))
	}

	g.types[name] = typ
}

func (g *Generator) saveRef(ref string, typ *ir.Type) {
	name, ok := typ.Name()
	if !ok {
		panic("unreachable")
	}

	if _, ok := g.refs.schemas[ref]; ok && !typ.IsGeneric() {
		panic(fmt.Sprintf("ref conflict: %q", ref))
	}

	if _, ok := g.types[name]; ok {
		panic(fmt.Sprintf("ref name conflict: %q", name))
	}

	g.refs.schemas[ref] = typ
	g.types[name] = typ
}
