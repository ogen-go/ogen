package ir

import "github.com/ogen-go/ogen/validate"

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
		if t.IsNumeric() && t.Primitive.IntValidation.Set() {
			return true
		}
		if t.Primitive.StringValidation.Set() {
			return true
		}
		return false
	case KindEnum:
		return true
	case KindSum:
		for _, s := range t.Sum.SumOf {
			if s.needValidation(path) {
				return true
			}
		}
		return false
	case KindAlias:
		return t.Alias.To.needValidation(path)
	case KindPointer:
		if t.Pointer.Semantic == NilInvalid {
			return true
		}
		return t.Pointer.To.needValidation(path)
	case KindGeneric:
		return t.Generic.Of.needValidation(path)
	case KindArray:
		if t.Array.Semantic == NilInvalid {
			return true
		}
		if t.Array.Validation.Set() {
			return true
		}
		return t.Array.Item.needValidation(path)
	case KindStruct:
		for _, f := range t.Struct.Fields {
			if f.Type.needValidation(path) {
				return true
			}
		}
		return false
	case KindStream:
		return false
	default:
		panic("unreachable")
	}
}
