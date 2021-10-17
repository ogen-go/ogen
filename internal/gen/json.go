package gen

import (
	"strings"

	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generatePrimitives() {
	for _, name := range []string{
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
	} {
		for _, v := range []struct {
			Optional bool
			Nil      bool
		}{
			{Optional: true, Nil: false},
			{Optional: false, Nil: true},
			{Optional: true, Nil: true},
		} {
			gt := &ast.Schema{
				Optional:  v.Optional,
				Nil:       v.Nil,
				Kind:      ast.KindPrimitive,
				Primitive: name,
			}
			switch name {
			case "uuid.UUID":
				gt.Format = "uuid"
			case "time.Duration":
				gt.Format = "duration"
			}
			if strings.Contains(name, "time.Time") {
				// The time.Time requires custom format, so setting special
				// value for this case.
				gt.Format = ast.FormatCustom
			}
			gt.Name = gt.GenericKind() + genericPostfix(name)
			g.generics = append(g.generics, gt)
		}
	}
}
