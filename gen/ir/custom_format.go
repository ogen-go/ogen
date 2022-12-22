package ir

import (
	"fmt"
	"reflect"
)

// ExternalType defines external type.
type ExternalType struct {
	// Pkg is name of the imported package.
	Pkg  string
	Type reflect.Type
}

// Go returns valid Go type for this ExternalType.
func (c ExternalType) Go() string {
	if c.Pkg == "" {
		// Primitive type.
		return c.Type.Name()
	}
	return fmt.Sprintf("%s.%s", c.Pkg, c.Type.Name())
}

// CustomFormat defines custom format type.
type CustomFormat struct {
	// Name is name of custom format.
	Name string
	// GoName is valid Go name for this custom format.
	GoName string
	// Type is type of custom format.
	Type ExternalType
	// JSON is JSON encoder/decoder for custom format.
	JSON ExternalType
	// Text is text encoder/decoder for custom format.
	Text ExternalType
}

// String returns string representation of CustomFormat.
func (c *CustomFormat) String() string {
	return c.GoName
}
