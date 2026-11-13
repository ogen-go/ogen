package gen

import "github.com/ogen-go/ogen/gen/ir"

// embedStructAliases marks aliases of struct types to be generated as
// structs embedding the aliased type, so its methods are promoted.
func (g *Generator) embedStructAliases() {
	for _, t := range g.tstorage.types {
		if t.Kind == ir.KindAlias && t.AliasTo != nil && t.AliasTo.Kind == ir.KindStruct {
			t.AliasEmbed = true
		}
	}
}
