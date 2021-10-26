package ir

// NilSemantic specifies nil value semantics.
type NilSemantic string

// Possible nil value semantics.
const (
	NilInvalid  NilSemantic = "invalid"  // nil is invalid
	NilOptional NilSemantic = "optional" // nil is "no value"
	NilNull     NilSemantic = "null"     // nil is null
)
