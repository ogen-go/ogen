package gen

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	ast "github.com/ogen-go/ogen/internal/ast2"
	"github.com/ogen-go/ogen/internal/ir"
	"golang.org/x/xerrors"
)

func fieldElem(s *ir.StructField) Elem {
	return Elem{
		SubElem: false,
		Field:   s.Tag,
		Type:    s.Type,
		Var:     fmt.Sprintf("s.%s", s.Name),
	}
}

// Elem variable helper for recursive array or object encoding.
type Elem struct {
	SubElem bool
	Field   string
	Type    *ir.Type
	Var     string
}

func (e Elem) NextVar() string {
	if !e.SubElem {
		return "elem"
	}
	return e.Var + "Elem"
}

// templateFuncs returns functions which used in templates.
func templateFuncs() template.FuncMap {
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

		// catent extra
		"pointer_elem": func(parent Elem) Elem {
			return Elem{
				Type:    parent.Type.PointerTo,
				SubElem: true,
				Var:     parent.NextVar(),
			}
		},
		"sub_array_elem": func(parent Elem, t *ir.Type) Elem {
			return Elem{
				Type:    t,
				SubElem: true,
				Var:     parent.NextVar(),
			}
		},
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
		"field_elem": fieldElem,
	}
}

//go:embed _template/*.tmpl
var templates embed.FS

// vendoredTemplates parses and returns vendored code generation templates.
func vendoredTemplates() *template.Template {
	tmpl := template.New("templates").Funcs(templateFuncs())
	tmpl = template.Must(tmpl.ParseFS(templates, "_template/*.tmpl"))
	return tmpl
}
