package ir

// Default represents default value.
type Default struct {
	Value interface{}
	Set   bool
}

// IsNil whether value is set, but null.
func (d Default) IsNil() bool {
	switch d.Value.(type) {
	case nil:
		return true
	default:
		return false
	}
}
