package gen

import (
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generatePrimitiveGenerics() {
	for _, primitive := range []string{
		"string",
		"int",
		"int32",
		"int64",
		"float32",
		"float64",
		"bool",
		"uuid.UUID",
		"time.Time",
		"time.Duration",
		"net.IP",
		"url.URL",
	} {
		for _, v := range []ast.GenericVariant{
			{Optional: true, Nullable: false},
			{Optional: false, Nullable: true},
			{Optional: true, Nullable: true},
		} {
			of := &ast.Schema{
				Kind:      ast.KindPrimitive,
				Primitive: primitive,
			}
			switch primitive {
			case "uuid.UUID":
				of.Format = "uuid"
			case "time.Duration":
				of.Format = "duration"
			case "net.IP":
				of.Format = "ip"
			case "url.URL":
				of.Format = "uri"
			}
			gt := ast.Generic(genericPostfix(of.Type()), of, v)
			g.schemas[gt.Name] = gt
		}
	}
}
