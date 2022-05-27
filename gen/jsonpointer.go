package gen

import "strings"

type jsonpointer struct {
	parts []string
}

func newJSONPointer(parts ...string) jsonpointer {
	return jsonpointer{parts: parts}
}

func (p jsonpointer) Append(parts ...string) jsonpointer {
	newParts := make([]string, 0, len(p.parts)+len(parts))
	newParts = append(newParts, p.parts...)
	newParts = append(newParts, parts...)
	return jsonpointer{parts: newParts}
}

var escapeReplacer = strings.NewReplacer(
	"/", "~1",
	"~", "~0",
)

func (p jsonpointer) String() string {
	escaped := make([]string, len(p.parts))
	for i, part := range p.parts {
		escaped[i] = escapeReplacer.Replace(part)
	}
	return strings.Join(escaped, "/")
}
