package ast

import (
	"strings"
)

type Generic struct {
	Schema
	Optional bool
	Nil      bool
}

func (g Generic) GenericKind() string {
	var b strings.Builder
	if g.Optional {
		b.WriteString("Optional")
	}
	if g.Nil {
		b.WriteString("Nil")
	}
	return b.String()
}
