package gen

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	"golang.org/x/xerrors"

	ast "github.com/ogen-go/ogen/internal/ast2"
	"github.com/ogen-go/ogen/internal/ir"
)

// Elem variable helper for recursive array or object encoding or decoding.
type Elem struct {
	SubElem bool
	Tag     ir.Tag
	Type    *ir.Type
	Var     string
}

// NextVar returns name of variable for decoding recursive call.
//
// Needed to make variable names unique.
func (e Elem) NextVar() string {
	if !e.SubElem {
		// No recursion, returning default name.
		return "elem"
	}
	return e.Var + "Elem"
}

// templateFunctions returns functions which used in templates.
func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"trim": strings.TrimSpace,
		"lower": func(v interface{}) string {
			switch v := v.(type) {
			case ast.ParameterLocation:
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
		"toString":   func(v interface{}) string { return fmt.Sprintf("%v", v) },
		"enumString": func(v interface{}) string {
			switch v := v.(type) {
			case string:
				return `"` + v + `"`
			case int, int8, int16, int32, int64, float32, float64, bool:
				return fmt.Sprintf("%v", v)
			case nil:
				return "nil"
			default:
				panic(fmt.Sprintf("unexpected type: %T", v))
			}
		},
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
				Type:    parent.Type.PointerTo,
				SubElem: true,
				Var:     parent.NextVar(),
			}
		},
		// Recursive array element (e.g. array of arrays).
		"sub_array_elem": func(parent Elem, t *ir.Type) Elem {
			return Elem{
				Type:    t,
				SubElem: true,
				Var:     parent.NextVar(),
			}
		},
		// Initial array element.
		"array_elem": func(t *ir.Type) Elem {
			return Elem{
				Type:    t,
				SubElem: true,
				Var:     "elem",
			}
		},
		"req_elem":        func(t *ir.Type) Elem { return Elem{Type: t, Var: "response"} },
		"req_decode_elem": func(t *ir.Type) Elem { return Elem{Type: t, Var: "request"} },
		"res_elem": func(i *ir.ResponseInfo) Elem {
			v := "response"
			if i.Default {
				v = v + ".Response"
			}
			return Elem{
				Type: i.Type,
				Var:  v,
			}
		},
		// Field of structure.
		"field_elem": func(s *ir.Field) Elem {
			return Elem{
				Tag:  s.Tag,
				Type: s.Type,
				Var:  fmt.Sprintf("s.%s", s.Name),
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
