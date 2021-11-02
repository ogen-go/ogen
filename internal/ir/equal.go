package ir

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
