package json

import "github.com/ogen-go/jx"

const (
	// Invalid invalid JSON element.
	Invalid = jx.Invalid
	// String JSON element "string".
	String = jx.String
	// Number JSON element 100 or 0.10.
	Number = jx.Number
	// Nil JSON element null.
	Nil = jx.Nil
	// Bool JSON element true or false.
	Bool = jx.Bool
	// Array JSON element [].
	Array = jx.Array
	// Object JSON element {}.
	Object = jx.Object
)
