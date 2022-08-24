package ir

// Default represents default value.
type Default struct {
	Value any
	Set   bool
}

// IsNil whether value is set, but null.
func (d Default) IsNil() bool {
	if !d.Set {
		return false
	}
	switch d.Value.(type) {
	case nil:
		return true
	default:
		return false
	}
}
