package ir

import (
	"regexp"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/jsonschema"
	"github.com/ogen-go/ogen/validate"
)

type Validators struct {
	String validate.String
	Int    validate.Int
	Array  validate.Array
	Object validate.Object
}

func (v *Validators) SetString(schema *jsonschema.Schema) (err error) {
	if schema.Pattern != "" {
		v.String.Regex, err = regexp.Compile(schema.Pattern)
		if err != nil {
			return errors.Wrap(err, "pattern")
		}
	}
	if schema.MaxLength != nil {
		v.String.SetMaxLength(int(*schema.MaxLength))
	}
	if schema.MinLength != nil {
		v.String.SetMinLength(int(*schema.MinLength))
	}
	if schema.Format == jsonschema.FormatEmail {
		v.String.Email = true
	}
	if schema.Format == jsonschema.FormatHostname {
		v.String.Hostname = true
	}
	return nil
}

func (v *Validators) SetInt(schema *jsonschema.Schema) {
	if schema.MultipleOf != nil {
		v.Int.SetMultipleOf(*schema.MultipleOf)
	}
	if schema.Maximum != nil {
		v.Int.SetMaximum(*schema.Maximum)
	}
	if schema.Minimum != nil {
		v.Int.SetMinimum(*schema.Minimum)
	}
	v.Int.MaxExclusive = schema.ExclusiveMaximum
	v.Int.MinExclusive = schema.ExclusiveMinimum
}

func (v *Validators) SetArray(schema *jsonschema.Schema) {
	if schema.MaxItems != nil {
		v.Array.SetMaxLength(int(*schema.MaxItems))
	}
	if schema.MinItems != nil {
		v.Array.SetMinLength(int(*schema.MinItems))
	}
}

func (v *Validators) SetObject(schema *jsonschema.Schema) {
	if schema.MaxProperties != nil {
		v.Object.SetMaxProperties(int(*schema.MaxProperties))
	}
	if schema.MinProperties != nil {
		v.Object.SetMinProperties(int(*schema.MinProperties))
	}
}

func (t *Type) NeedValidation() bool {
	return t.needValidation(&walkpath{})
}

func (t *Type) needValidation(path *walkpath) (result bool) {
	if t == nil {
		return false
	}

	if path.has(t) {
		return false
	}

	path = path.append(t)

	switch t.Kind {
	case KindPrimitive:
		if t.IsFloat() {
			// NaN, Inf.
			return true
		}
		if t.IsNumeric() && t.Validators.Int.Set() {
			return true
		}
		if t.Validators.String.Set() {
			return true
		}
		return false
	case KindEnum:
		return true
	case KindSum:
		for _, s := range t.SumOf {
			if s.needValidation(path) {
				return true
			}
		}
		return false
	case KindAlias:
		return t.AliasTo.needValidation(path)
	case KindPointer:
		if t.NilSemantic == NilInvalid {
			return true
		}
		return t.PointerTo.needValidation(path)
	case KindGeneric:
		return t.GenericOf.needValidation(path)
	case KindArray:
		if t.NilSemantic == NilInvalid {
			return true
		}
		if t.Validators.Array.Set() {
			return true
		}
		return t.Item.needValidation(path)
	case KindStruct:
		for _, f := range t.Fields {
			if f.Type.needValidation(path) {
				return true
			}
		}
		return false
	case KindMap:
		if t.Validators.Object.Set() {
			return true
		}
		return t.Item.needValidation(path)
	case KindStream, KindAny:
		// FIXME(tdakkota): try to validate Any.
		return false
	default:
		panic("unreachable")
	}
}
