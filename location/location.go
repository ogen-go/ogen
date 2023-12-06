// Package location provides utilities to track values over the spec.
package location

// Pointer is a location of a value.
type Pointer struct {
	// Source is the File where the value is located.
	Source File
	// Locator stores the Position of a value.
	Locator Locator
}

// File returns the File where the value is located.
func (p Pointer) File() File {
	return p.Source
}

// Position returns the position of the value if it is set.
func (p Pointer) Position() (Position, bool) {
	return p.Locator.Position()
}

// Key tries to find the child node using given key and returns its pointer.
//
// See Key method of Locator.
func (p Pointer) Key(key string) (ptr Pointer) {
	l := p.Locator.Key(key)
	if !l.set {
		return p
	}
	return Pointer{
		Locator: l,
		Source:  p.Source,
	}
}

// Field tries to find the child node using given key and returns its pointer.
//
// See Field method of Locator.
func (p Pointer) Field(key string) (ptr Pointer) {
	l := p.Locator.Field(key)
	if !l.set {
		return p
	}
	return Pointer{
		Locator: l,
		Source:  p.Source,
	}
}

// Index tries to find the child node using given index and returns its pointer.
//
// See Index method of Locator.
func (p Pointer) Index(idx int) (ptr Pointer) {
	l := p.Locator.Index(idx)
	if !l.set {
		return p
	}
	return Pointer{
		Locator: l,
		Source:  p.Source,
	}
}
