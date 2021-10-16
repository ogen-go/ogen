package gen

import (
	"embed"
	"fmt"
	"io/fs"
	"reflect"
	"strings"

	"github.com/ogen-go/ogen/internal/ast"
	"github.com/open2b/scriggo"
	"github.com/open2b/scriggo/native"
	"golang.org/x/xerrors"
)

// templateFuncs returns functions which used in templates.
func templateFuncs() native.Declarations {
	return native.Declarations{
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

func vendoredTemplates() *scriggo.Files {
	matches, err := fs.Glob(templates, "_template/*.tmpl")
	if err != nil {
		panic(err)
	}

	files := make(scriggo.Files)
	for _, fileName := range matches {
		b, err := fs.ReadFile(templates, fileName)
		if err != nil {
			panic(err)
		}

		files[strings.TrimPrefix(fileName, "_template/")] = b
	}
	return &files
}

func astPkg() native.Package {
	decs := make(native.Declarations)
	decs["Method"] = reflect.TypeOf((*ast.Method)(nil)).Elem()
	decs["Parameter"] = reflect.TypeOf((*ast.Parameter)(nil)).Elem()
	decs["ParameterLocation"] = reflect.TypeOf((*ast.ParameterLocation)(nil)).Elem()
	decs["RequestBody"] = reflect.TypeOf((*ast.RequestBody)(nil)).Elem()
	decs["MethodResponse"] = reflect.TypeOf((*ast.MethodResponse)(nil)).Elem()
	decs["Response"] = reflect.TypeOf((*ast.Response)(nil)).Elem()
	decs["Schema"] = reflect.TypeOf((*ast.Schema)(nil)).Elem()
	decs["PathPart"] = reflect.TypeOf((*ast.PathPart)(nil)).Elem()
	decs["Interface"] = reflect.TypeOf((*ast.Interface)(nil)).Elem()

	return native.Package{
		Name:         "ast",
		Declarations: decs,
	}
}
