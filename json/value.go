package json

import "github.com/go-faster/jx"

const (
	// Invalid invalid JSON element.
	Invalid = jx.Invalid
	// String JSON element "string".
	String = jx.String
	// Number JSON element 100 or 0.10.
	Number = jx.Number
	// Null JSON element null.
	Null = jx.Null
	// Bool JSON element true or false.
	Bool = jx.Bool
	// Array JSON element [].
	Array = jx.Array
	// Object JSON element {}.
	Object = jx.Object
)
