package gen

import (
	"embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/naming"
)

// OperationElem is variable name for generating per-operation functions.
type OperationElem struct {
	// Operation is the operation.
	Operation *ir.Operation
	// Config is the template configuration.
	Config TemplateConfig
}

// RouterElem is variable helper for router generation.
type RouterElem struct {
	// ParameterIndex is index of parameter of this route part.
	ParameterIndex int
	Route          *RouteNode
}

// DefaultElem is variable helper for setting default values.
type DefaultElem struct {
	// Type is type of this DefaultElem.
	Type *ir.Type
	// Var is decoding/encoding variable Go name (obj) or selector (obj.Field).
	Var string
	// Default is default value to set.
	Default ir.Default
}

// Elem is variable helper for recursive array or object encoding or decoding.
type Elem struct {
	// Sub whether this Elem has parent Elem.
	Sub bool
	// Type is type of this Elem.
	Type *ir.Type
	// Var is decoding/encoding variable Go name (obj) or selector (obj.Field).
	Var string
	// Tag contains info about field tags, if any.
	Tag ir.Tag
	// First whether this field is first.
	First bool
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

type ResponseElem struct {
	Response *ir.Response
	Ptr      bool
}

// templateFunctions returns functions which used in templates.
func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"errorf": func(format string, args ...any) (any, error) {
			return nil, errors.Errorf(format, args...)
		},
		"pascalSpecial": pascalSpecial,
		"camelSpecial":  camelSpecial,
		"capitalize":    naming.Capitalize,
		"upper":         strings.ToUpper,

		// Helpers for recursive encoding and decoding.
		"elem": func(t *ir.Type, v string) Elem {
			return Elem{
				Type: t,
				Var:  v,
			}
		},
		"pointer_elem": func(parent Elem) Elem {
			return Elem{
				Type: parent.Type.PointerTo,
				Sub:  true,
				Var:  parent.NextVar(),
			}
		},
		// Recursive array element (e.g. array of arrays).
		"sub_array_elem": func(parent Elem, t *ir.Type) Elem {
			return Elem{
				Type: t,
				Sub:  true,
				Var:  parent.NextVar(),
			}
		},
		// Initial array element.
		"array_elem": func(t *ir.Type) Elem {
			return Elem{
				Type: t,
				Sub:  true,
				Var:  "elem",
			}
		},
		"map_elem": func(t *ir.Type) Elem {
			return Elem{
				Type: t,
				Sub:  true,
				Var:  "elem",
			}
		},
		// Field of structure.
		"field_elem": func(f *ir.Field) Elem {
			return Elem{
				Type: f.Type,
				Var:  fmt.Sprintf("s.%s", f.Name),
				Tag:  f.Tag,
			}
		},
		"first_field_elem": func(f *ir.Field) Elem {
			return Elem{
				Type:  f.Type,
				Var:   fmt.Sprintf("s.%s", f.Name),
				Tag:   f.Tag,
				First: true,
			}
		},
		"response_elem": func(r *ir.Response, ptr bool) ResponseElem {
			return ResponseElem{
				Response: r,
				Ptr:      ptr,
			}
		},
		"router_elem": func(child *RouteNode, currentIdx int) RouterElem {
			if child.IsParam() {
				currentIdx++
			}
			return RouterElem{
				ParameterIndex: currentIdx,
				Route:          child,
			}
		},
		"default_elem": func(t *ir.Type, v string, value ir.Default) DefaultElem {
			return DefaultElem{
				Type:    t,
				Var:     v,
				Default: value,
			}
		},
		"sub_default_elem": func(t *ir.Type, v string, val any) DefaultElem {
			return DefaultElem{
				Type: t,
				Var:  v,
				Default: ir.Default{
					Value: val,
					Set:   true,
				},
			}
		},
		"op_elem": func(op *ir.Operation, cfg TemplateConfig) OperationElem {
			return OperationElem{
				Operation: op,
				Config:    cfg,
			}
		},
		"ir_media": func(e ir.Encoding, t *ir.Type) ir.Media {
			return ir.Media{
				Encoding: e,
				Type:     t,
			}
		},
		"print_go": ir.PrintGoValue,
		// We use any to prevent template type matching errors
		// for type aliases (e.g. for quoting ir.ContentType).
		"quote": func(v any) string {
			// Fast path for string.
			if s, ok := v.(string); ok {
				return strconv.Quote(s)
			}
			return fmt.Sprintf("%q", v)
		},
		"backquote": func(v any) string {
			// Fast path for string.
			if s, ok := v.(string); ok && strconv.CanBackquote(s) {
				return "`" + s + "`"
			}
			return fmt.Sprintf("%#q", v)
		},
		"times": func(n int) []struct{} {
			return make([]struct{}, n)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"div": func(a, b int) int {
			return a / b
		},
		"mod": func(a, b int) int {
			return a % b
		},
		"isObjectParam":     isObjectParam,
		"paramObjectFields": paramObjectFields,
	}
}

//go:embed _template/*
var templates embed.FS

var _templates struct {
	sync.Once
	val *template.Template
}

// vendoredTemplates parses and returns vendored code generation templates.
func vendoredTemplates() *template.Template {
	_templates.Do(func() {
		tmpl := template.New("templates").Funcs(templateFunctions())
		tmpl = template.Must(tmpl.ParseFS(templates,
			"_template/*.tmpl",
			"_template/*/*.tmpl",
		))
		_templates.val = tmpl
	})
	return _templates.val
}

func isObjectParam(p *ir.Parameter) bool {
	if p.Spec != nil && p.Spec.Content != nil {
		// "content" encoding.
		return false
	}
	typ := p.Type
	if typ.IsGeneric() {
		typ = typ.GenericOf
	}

	return typ.IsStruct()
}

func paramObjectFields(typ *ir.Type) string {
	if typ.IsGeneric() {
		typ = typ.GenericOf
	}

	if !typ.IsStruct() {
		return "nil"
	}

	fields := make([]string, 0, len(typ.Fields))
	for _, f := range typ.Fields {
		if f.Spec == nil {
			continue
		}

		req := "false"
		if f.Spec.Required {
			req = "true"
		}

		fields = append(fields, "{Name:\""+f.Spec.Name+"\",Required:"+req+"}")
	}

	return "[]uri.QueryParameterObjectField{" + strings.Join(fields, ",") + "}"
}
