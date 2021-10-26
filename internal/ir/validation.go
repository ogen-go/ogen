package ir

import "github.com/ogen-go/ogen/validate"

type Validators struct {
	String validate.String
	Int    validate.Int
	Array  validate.Array
}

func (t *Type) NeedValidation() bool {
	return t.needValidation(map[*Type]struct{}{})
}

func (t *Type) needValidation(visited map[*Type]struct{}) (result bool) {
	if t == nil {
		return false
	}

	if _, ok := visited[t]; ok {
		return false
	}

	visited[t] = struct{}{}

	switch t.Kind {
	case KindPrimitive:
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
			if s.needValidation(visited) {
				return true
			}
		}
		return false
	case KindAlias:
		return t.AliasTo.needValidation(visited)
	case KindPointer:
		if t.NilSemantic == NilInvalid {
			return true
		}
		return t.PointerTo.needValidation(visited)
	case KindGeneric:
		return t.GenericOf.needValidation(visited)
	case KindArray:
		if t.NilSemantic == NilInvalid {
			return true
		}
		if t.Validators.Array.Set() {
			return true
		}
		// Prevent infinite recursion.
		if t.Item == t {
			return false
		}
		return t.Item.needValidation(visited)
	case KindStruct:
		for _, f := range t.Fields {
			if f.Type.needValidation(visited) {
				return true
			}
		}
		return false
	default:
		panic("unreachable")
	}
}
