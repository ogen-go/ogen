package gen

import (
	"fmt"

	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) saveIface(t *ir.Type) {
	if !t.Is(ir.KindInterface) {
		panic(unreachable(t))
	}

	if _, ok := g.interfaces[t.Name]; ok {
		panic(fmt.Sprintf("interface name conflict: %q", t.Name))
	}

	g.interfaces[t.Name] = t
}
