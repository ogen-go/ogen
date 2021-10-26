package gen

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	"golang.org/x/xerrors"
)

// templateFuncs returns functions which used in templates.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"trim":       strings.TrimSpace,
		"lower":      strings.ToLower,
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
