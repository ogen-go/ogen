package ir

import (
	"strconv"

	"github.com/ogen-go/ogen/jsonschema"
)

// Tag of Field.
type Tag struct {
	JSON string // json tag, empty for none
}

// EscapedJSON returns quoted and escaped JSON tag.
func (t Tag) EscapedJSON() string {
	return strconv.Quote(t.JSON)
}

// Field of structure.
type Field struct {
	Name string
	Type *Type
	Tag  Tag
	Spec *jsonschema.Property
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
	if f.Spec == nil {
		return nil
	}
	return prettyDoc(f.Spec.Description)
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
	for _, f := range t.Fields {
		if val := f.Default(); val.Set {
			return true
		}
	}
	return false
}
