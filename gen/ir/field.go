package ir

import (
	"fmt"
	"strconv"

	"github.com/ogen-go/ogen/internal/xslices"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

// Tag of Field.
type Tag struct {
	JSON string             // json tag, empty for none
	Form *openapi.Parameter // query form parameter
}

// EscapedJSON returns quoted and escaped JSON tag.
func (t Tag) EscapedJSON() string {
	return strconv.Quote(t.JSON)
}

// InlineField defines how to inline field.
type InlineField int

const (
	InlineNone InlineField = iota
	InlineAdditional
	InlinePattern
)

// Field of structure.
type Field struct {
	// Go Name of field.
	Name string
	// Type of field.
	Type *Type
	// JSON tag. May be empty.
	Tag Tag
	// Whether field is inlined map (i.e. additionalProperties, patternProperties).
	Inline InlineField
	// Spec is property schema. May be nil.
	Spec *jsonschema.Property
}

// StructTag returns struct tag for field.
func (f Field) StructTag() string {
	var (
		tag        = f.Tag.JSON
		isOptional = f.Type.GenericVariant.Optional ||
			(f.Spec != nil && !f.Spec.Required)
	)
	if tag == "" && !isOptional {
		return ""
	}

	if isOptional {
		tag += ",omitempty"
	}
	s := "json:" + strconv.Quote(tag)
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// ValidationName returns name for FieldError.
func (f Field) ValidationName() string {
	if f.Spec != nil {
		return f.Spec.Name
	}
	return f.Name
}

// Default returns default value of this field, if it is set.
func (f Field) Default() Default {
	var schema *jsonschema.Schema
	if spec := f.Spec; spec != nil {
		schema = spec.Schema
	}
	if schema != nil {
		return Default{
			Value: schema.Default,
			Set:   schema.DefaultSet,
		}
	}
	if typ := f.Type; typ != nil {
		return typ.Default()
	}
	return Default{}
}

// GoDoc returns field godoc.
func (f Field) GoDoc() []string {
	s := f.Spec
	if s == nil {
		if f.Inline == InlinePattern {
			return []string{fmt.Sprintf("Pattern: %q.", f.Type.MapPattern)}
		}
		return nil
	}

	var notice string
	if sch := s.Schema; sch != nil && sch.Deprecated {
		notice = "Deprecated: schema marks this property as deprecated."
	}

	return prettyDoc(s.Description, notice)
}

// DefaultFields returns fields with default values.
func (t Type) DefaultFields() (r []*Field) {
	for _, f := range t.Fields {
		if val := f.Default(); val.Set {
			r = append(r, f)
		}
	}
	return r
}

// HasDefaultFields whether type has fields with default values.
func (t Type) HasDefaultFields() bool {
	return xslices.ContainsFunc(t.Fields, func(f *Field) bool {
		return f.Default().Set
	})
}

func (t Type) parameters(keep func(t *Type) bool) (params []Parameter) {
	if !t.IsStruct() {
		panic(fmt.Sprintf("unreachable: %s", t))
	}
	for _, f := range t.Fields {
		if !keep(f.Type) {
			continue
		}
		params = append(params, Parameter{
			Name: f.Name,
			Type: f.Type,
			Spec: f.Tag.Form,
		})
	}
	return params
}

func (t Type) FormParameters() (params []Parameter) {
	return t.parameters(func(t *Type) bool {
		return !t.HasFeature("multipart-file")
	})
}

func (t Type) FileParameters() (params []Parameter) {
	return t.parameters(func(t *Type) bool {
		return t.HasFeature("multipart-file")
	})
}
