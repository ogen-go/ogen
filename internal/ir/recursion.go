package ir

// RecursiveTo reports whether target type is recursive to current.
func (t *Type) RecursiveTo(to *Type) bool {
	if to.Is(KindPrimitive, KindArray) {
		return false
	}
	return t.recurse(t, to)
}

// Equal reports whether two types are equal.
func (t *Type) Equal(target *Type) bool {
	if t.Kind != target.Kind {
		return false
	}
	if t.Primitive != target.Primitive {
		return false
	}
	return t.Go() == target.Go()
}

func (t *Type) recurse(parent, target *Type) bool {
	if target.Equal(parent) {
		return true
	}
	for _, f := range target.Fields {
		if parent.RecursiveTo(f.Type) {
			return true
		}
	}
	if t.GenericOf != nil {
		return t.GenericOf.RecursiveTo(target)
	}
	if target.GenericOf != nil {
		return t.recurse(parent, target.GenericOf)
	}
	return false
}
