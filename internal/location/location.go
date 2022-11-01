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
