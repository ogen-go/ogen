package ir

import (
	"github.com/ogen-go/ogen/validate"
)

type Validators struct {
	String validate.String
	Int    validate.Int
	Array  validate.Array
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
		return t.Item.needValidation(path)
	case KindStream:
		return false
	default:
		panic("unreachable")
	}
}
