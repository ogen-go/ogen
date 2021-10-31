package gen

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

// Elem variable helper for recursive array or object encoding or decoding.
type Elem struct {
	Sub  bool // true if Elem has parent Elem
	Type *ir.Type
	Var  string
	Tag  ir.Tag
	More bool
}

// NextVar returns name of variable for decoding recursive call.
//
// Needed to make variable names unique.
func (e Elem) NextVar() string {
	if !e.Sub {
		// No recursion, returning default name.
		return "elem"
	}
	return e.Var + "Elem"
}

// templateFunctions returns functions which used in templates.
func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"inc":  func(i int) int { return i + 1 },
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
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, xerrors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, xerrors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"sprintf": fmt.Sprintf,

		// Helpers for recursive encoding and decoding.
		"pointer_elem": func(parent Elem) Elem {
			return Elem{
				Type: parent.Type.PointerTo,
				Sub:  true,
				Var:  parent.NextVar(),
				More: true,
			}
		},
		// Recursive array element (e.g. array of arrays).
		"sub_array_elem": func(parent Elem, t *ir.Type) Elem {
			return Elem{
				Type: t,
				Sub:  true,
				Var:  parent.NextVar(),
				More: true,
			}
		},
		// Initial array element.
		"array_elem": func(t *ir.Type) Elem {
			return Elem{
				Type: t,
				Sub:  true,
				Var:  "elem",
				More: true,
			}
		},
		"req_elem":     func(t *ir.Type) Elem { return Elem{Type: t, Var: "response", More: true} },
		"req_dec_elem": func(t *ir.Type) Elem { return Elem{Type: t, Var: "request", More: true} },
		"req_enc_elem": func(t *ir.Type) Elem { return Elem{Type: t, Var: "req", More: true} },
		"res_elem": func(i *ir.ResponseInfo) Elem {
			v := "response"
			if i.Default {
				v = v + ".Response"
			}
			return Elem{
				Type: i.Type,
				Var:  v,
				More: true,
			}
		},
		// Field of structure.
		"field_elem": func(f *ir.Field) Elem {
			return Elem{
				Type: f.Type,
				Var:  fmt.Sprintf("s.%s", f.Name),
				Tag:  f.Tag,
				More: true,
			}
		},
		// Element of sum type.
		"sum_elem": func(t *ir.Type) Elem {
			return Elem{
				Type: t,
				Var:  fmt.Sprintf("s.%s", t.Name),
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
