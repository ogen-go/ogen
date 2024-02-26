package location

// File is a source file.
type File struct {
	// Name is the file name.
	Name string
	// Source is the path or URL of the file.
	Source string
	// Lines stores newline offsets.
	Lines Lines
}

// HumanName returns human-friendly name for this File.
func (f File) HumanName() string {
	if n := f.Name; n != "" {
		return n
	}
	return f.Source
}

// IsZero returns true if file has zero value.
func (f File) IsZero() bool {
	s := struct {
		Name   string
		Source string
		Lines  Lines
	}(f)
	// File is not useful if lines is empty.
	return (s.Name == "" && s.Source == "") || s.Lines.IsZero()
}

// NewFile creates a new File.
//
// Do not modify the data after calling this function, Lines will point to it.
func NewFile(name, source string, data []byte) File {
	f := File{
		Name:   name,
		Source: source,
	}
	f.Lines.Collect(data)
	return f
}
