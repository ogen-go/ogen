package gen

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

// Elem is a template helper.
// Used to pass type info and variable name through recursive templates
// (json encoding/decoding, uri encoding/decoding, validation).
type Elem struct {
	Type *ir.Type
	Var  string
}

type ResponseElem struct {
	Response *ir.StatusResponse
	Ptr      bool
}

// JXElem is a wrapper around Elem.
type JXElem struct {
	Field string // Optional JSON field name.
	Elem
}

// templateFunctions returns functions which used in templates.
func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"trim": strings.TrimSpace,
		"lower": func(v interface{}) string {
			switch v := v.(type) {
			case oas.ParameterLocation:
				return strings.ToLower(string(v))
			case string:
				return strings.ToLower(v)
			default:
				panic(fmt.Sprintf("unexpected value: %T", v))
			}
		},
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"pascalMP":   pascalMP,

		// Helpers for recursive encoding and decoding.
		"elem": func(t *ir.Type, v string) Elem {
			return Elem{
				Type: t,
				Var:  v,
			}
		},
		"jx_elem": func(t *ir.Type, v, f string) JXElem {
			return JXElem{
				Field: f,
				Elem: Elem{
					Type: t,
					Var:  v,
				},
			}
		},
		"resp_elem": func(r *ir.StatusResponse, ptr bool) ResponseElem {
			return ResponseElem{
				Response: r,
				Ptr:      ptr,
			}
		},
	}
}

//go:embed _template/*.tmpl
var templates embed.FS

// vendoredTemplates parses and returns vendored code generation templates.
func vendoredTemplates() *template.Template {
	tmpl := template.New("templates").Funcs(templateFunctions())
	tmpl = template.Must(tmpl.ParseFS(templates, "_template/*.tmpl"))
	return tmpl
}
