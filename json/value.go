package json

import (
	"github.com/ogen-go/jir"
)

const (
	// Invalid invalid JSON element.
	Invalid = jir.Invalid
	// String JSON element "string".
	String = jir.String
	// Number JSON element 100 or 0.10.
	Number = jir.Number
	// Nil JSON element null.
	Nil = jir.Nil
	// Bool JSON element true or false.
	Bool = jir.Bool
	// Array JSON element [].
	Array = jir.Array
	// Object JSON element {}.
	Object = jir.Object
)
